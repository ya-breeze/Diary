package tasks

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"database/sql"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"time"

	_ "github.com/mattn/go-sqlite3" // register sqlite3 driver
	"github.com/ya-breeze/diary.be/pkg/config"
)

const (
	backupDateFormat = "2006-01-02"
	backupPrefix     = "diary-backup-"
	backupSuffix     = ".tar.gz"
)

// BackupTask creates daily .tar.gz backups of the database and assets.
type BackupTask struct {
	logger   *slog.Logger
	cfg      *config.Config
	interval time.Duration
}

func NewBackupTask(logger *slog.Logger, cfg *config.Config) *BackupTask {
	interval := 24 * time.Hour
	if cfg.BackupInterval != "" {
		if d, err := time.ParseDuration(cfg.BackupInterval); err == nil {
			interval = d
		} else {
			logger.Warn("Invalid backup_interval, using 24h", "value", cfg.BackupInterval, "error", err)
		}
	}
	return &BackupTask{logger: logger, cfg: cfg, interval: interval}
}

// Start launches the background goroutine. It waits 30s on startup (matching CheckerTask),
// then runs on the ticker.
func (t *BackupTask) Start(ctx context.Context) {
	go func() {
		select {
		case <-time.After(30 * time.Second):
		case <-ctx.Done():
			return
		}
		t.run()
		ticker := time.NewTicker(t.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				t.run()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (t *BackupTask) run() {
	today := time.Now().Format(backupDateFormat)
	backupDir := filepath.Join(t.cfg.DataPath, config.BackupsDirName)

	if err := os.MkdirAll(backupDir, 0o750); err != nil {
		t.logger.Error("backup: failed to create backup directory", "error", err)
		return
	}

	archivePath := filepath.Join(backupDir, backupPrefix+today+backupSuffix)
	if _, err := os.Stat(archivePath); err == nil {
		t.logger.Info("backup: today's backup already exists, skipping", "date", today)
		return
	}

	t.logger.Info("backup: starting", "date", today)

	tmpDB := archivePath + ".db.tmp"
	defer os.Remove(tmpDB) //nolint:errcheck

	if err := vacuumInto(filepath.Join(t.cfg.DataPath, config.DBFilename), tmpDB); err != nil {
		t.logger.Error("backup: VACUUM INTO failed", "error", err)
		return
	}

	tmpArchive := archivePath + ".tmp"
	defer os.Remove(tmpArchive) //nolint:errcheck

	assetsDir := filepath.Join(t.cfg.DataPath, config.AssetsDirName)
	if err := createArchive(tmpArchive, tmpDB, assetsDir); err != nil {
		t.logger.Error("backup: failed to create archive", "error", err)
		return
	}

	if err := os.Rename(tmpArchive, archivePath); err != nil {
		t.logger.Error("backup: failed to finalize archive", "error", err)
		return
	}

	t.logger.Info("backup: completed", "file", filepath.Base(archivePath))

	if err := t.pruneBackups(backupDir); err != nil {
		t.logger.Error("backup: pruning failed", "error", err)
	}
}

// vacuumInto executes VACUUM INTO using a fresh database/sql connection.
// This produces an atomic, consistent copy without requiring GORM access.
func vacuumInto(src, dst string) error {
	db, err := sql.Open("sqlite3", src)
	if err != nil {
		return fmt.Errorf("open source db: %w", err)
	}
	defer db.Close() //nolint:errcheck

	if _, err := db.Exec("VACUUM INTO ?", dst); err != nil {
		return fmt.Errorf("vacuum into: %w", err)
	}
	return nil
}

func createArchive(archivePath, tmpDB, assetsDir string) error {
	f, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("create archive: %w", err)
	}
	defer f.Close() //nolint:errcheck

	gw := gzip.NewWriter(f)
	defer gw.Close() //nolint:errcheck
	tw := tar.NewWriter(gw)
	defer tw.Close() //nolint:errcheck

	if err := addFileToTar(tw, tmpDB, config.DBFilename); err != nil {
		return fmt.Errorf("add db: %w", err)
	}
	if _, err := os.Stat(assetsDir); err == nil {
		if err := addDirToTar(tw, assetsDir, config.AssetsDirName); err != nil {
			return fmt.Errorf("add assets: %w", err)
		}
	}
	return nil
}

func addFileToTar(tw *tar.Writer, srcPath, archiveName string) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck

	info, err := f.Stat()
	if err != nil {
		return err
	}
	hdr := &tar.Header{
		Name:    archiveName,
		Size:    info.Size(),
		Mode:    int64(info.Mode()),
		ModTime: info.ModTime(),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	_, err = io.Copy(tw, f)
	return err
}

func addDirToTar(tw *tar.Writer, srcDir, archiveBase string) error {
	return filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(srcDir, path)
		archivePath := filepath.Join(archiveBase, rel)

		info, err := d.Info()
		if err != nil {
			return err
		}
		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = archivePath
		if d.IsDir() {
			hdr.Name += "/"
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close() //nolint:errcheck
		_, err = io.Copy(tw, f)
		return err
	})
}

func (t *BackupTask) pruneBackups(backupDir string) error {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("read backup dir: %w", err)
	}

	var names []string
	for _, e := range entries {
		n := e.Name()
		if !e.IsDir() &&
			len(n) > len(backupPrefix)+len(backupSuffix) &&
			n[:len(backupPrefix)] == backupPrefix &&
			n[len(n)-len(backupSuffix):] == backupSuffix {
			names = append(names, n)
		}
	}

	sort.Strings(names) // lexicographic == chronological for YYYY-MM-DD names

	maxCount := t.cfg.BackupMaxCount
	if maxCount <= 0 {
		maxCount = 10
	}

	for len(names) > maxCount {
		oldest := names[0]
		names = names[1:]
		if err := os.Remove(filepath.Join(backupDir, oldest)); err != nil {
			t.logger.Warn("backup: failed to delete old backup", "file", oldest, "error", err)
		} else {
			t.logger.Info("backup: deleted old backup", "file", oldest)
		}
	}
	return nil
}
