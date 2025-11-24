package database

import (
	"context"
	"log/slog"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SlogGormLogger struct {
	logger  *slog.Logger
	verbose bool
	level   logger.LogLevel
}

func (l *SlogGormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.level = level
	return &newLogger
}

func (l *SlogGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= logger.Info {
		l.logger.InfoContext(ctx, msg, data...)
	}
}

func (l *SlogGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= logger.Warn {
		l.logger.WarnContext(ctx, msg, data...)
	}
}

func (l *SlogGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= logger.Error {
		l.logger.ErrorContext(ctx, msg, data...)
	}
}

func (l *SlogGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if !l.verbose {
		return
	}
	elapsed := time.Since(begin)
	sql, rows := fc()
	if err != nil {
		l.logger.ErrorContext(ctx, "trace", "sql", sql, "rows", rows, "elapsed", elapsed, "error", err)
	} else {
		l.logger.InfoContext(ctx, "trace", "sql", sql, "rows", rows, "elapsed", elapsed)
	}
}

func openSqlite(l *slog.Logger, dbPath string, verbose bool) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: (&SlogGormLogger{logger: l, verbose: verbose}).LogMode(logger.Warn),
	})
}
