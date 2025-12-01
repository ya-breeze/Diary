package webapp

import (
	"fmt"
	"html/template"
	"log/slog"
	"maps"
	"math"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/database"
	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
	"github.com/ya-breeze/diary.be/pkg/utils"
)

type WebAppRouter struct {
	commit       string
	logger       *slog.Logger
	cfg          *config.Config
	db           database.Storage
	cookies      *sessions.CookieStore
	authService  goserver.AuthAPIService
	itemsService goserver.ItemsAPIService
}

func NewWebAppRouter(
	controllers goserver.CustomControllers, commit string, logger *slog.Logger, cfg *config.Config, db database.Storage,
) *WebAppRouter {
	return &WebAppRouter{
		commit:       commit,
		logger:       logger,
		cfg:          cfg,
		db:           db,
		cookies:      sessions.NewCookieStore([]byte(cfg.SessionSecret)),
		authService:  controllers.AuthAPIService,
		itemsService: controllers.ItemsAPIService,
	}
}

func (r *WebAppRouter) Routes() goserver.Routes {
	res := goserver.Routes{}
	merge := func(m goserver.Routes) { maps.Copy(res, m) }
	merge(r.routesCore())
	merge(r.routesUploads())
	merge(r.routesStatic())
	return res
}

func (r *WebAppRouter) routesCore() goserver.Routes {
	return goserver.Routes{
		"RootPath":  {Method: "GET", Pattern: "/", HandlerFunc: r.homeHandler},
		"Login":     {Method: "POST", Pattern: "/web/login", HandlerFunc: r.loginHandler},
		"Logout":    {Method: "GET", Pattern: "/web/logout", HandlerFunc: r.logoutHandler},
		"AboutPath": {Method: "GET", Pattern: "/web/about", HandlerFunc: r.aboutHandler},
		"Search":    {Method: "GET", Pattern: "/web/search", HandlerFunc: r.searchHandler},
		"Edit":      {Method: "GET", Pattern: "/web/edit", HandlerFunc: r.editHandler},
		"Save":      {Method: "POST", Pattern: "/web/edit", HandlerFunc: r.saveHandler},
	}
}

func (r *WebAppRouter) routesUploads() goserver.Routes {
	return goserver.Routes{
		"Upload":      {Method: "POST", Pattern: "/web/upload", HandlerFunc: r.uploadHandler},
		"UploadBatch": {Method: "POST", Pattern: "/web/upload-batch", HandlerFunc: r.uploadBatchHandler},
	}
}

func (r *WebAppRouter) routesStatic() goserver.Routes {
	return goserver.Routes{
		"Assets": {Method: "GET", Pattern: "/web/assets/{rest:.*}", HandlerFunc: r.assetsHandler},
		"Static": {Method: "GET", Pattern: "/web/static/{rest:.*}", HandlerFunc: r.staticHandler},
	}
}

func (r *WebAppRouter) loadTemplates() (*template.Template, error) {
	tmpl, err := template.New("").Funcs(template.FuncMap{
		"formatTime": utils.FormatTime,
		"decrease": func(i int) int {
			return i - 1
		},
		"money": func(num float64) float64 {
			return math.Round(num*100) / 100
		},
		"timestamp": func(t time.Time) int64 {
			return t.Unix()
		},
		"lastMonth": func(t time.Time) time.Time {
			return time.Date(t.Year(), t.Month()-1, 1, 0, 0, 0, 0, t.Location())
		},
		"addMonths": func(t time.Time, num int) time.Time {
			return time.Date(t.Year(), t.Month()+time.Month(num), 1, 0, 0, 0, 0, t.Location())
		},
		"addQueryParam": func(rawURL string, key string, value any) (string, error) {
			u, err := url.Parse(rawURL)
			if err != nil {
				return "", err
			}
			q := u.Query()
			q.Set(key, fmt.Sprintf("%v", value))
			u.RawQuery = q.Encode()
			return u.String(), nil
		},
		"truncate": func(s string, length int) string {
			if len(s) <= length {
				return s
			}
			return s[:length]
		},
		// snippet collapses whitespace and trims trailing underscores/asterisks to keep preview inline
		"snippet": func(s string, length int) string {
			// Basic truncate first
			if len(s) > length {
				s = s[:length]
			}
			// Normalize whitespace to single spaces
			s = strings.ReplaceAll(s, "\r", " ")
			s = strings.ReplaceAll(s, "\n", " ")
			s = strings.ReplaceAll(s, "\t", " ")
			// Collapse multiple spaces
			for strings.Contains(s, "  ") {
				s = strings.ReplaceAll(s, "  ", " ")
			}
			// Trim spaces
			s = strings.TrimSpace(s)
			// Trim dangling markdown emphasis/underscore fragments at the end
			s = strings.TrimRight(s, "_*")
			return s
		},
	}).ParseGlob(filepath.Join("webapp", "templates", "*.tpl"))
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}
