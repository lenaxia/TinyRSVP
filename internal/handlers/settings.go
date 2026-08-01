package handlers

import (
	"html/template"
	"net/http"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/config"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

const redactedPlaceholder = "••••••••"

type SettingsView struct {
	Server      ServerSettingsView
	Database    DatabaseSettingsView
	Auth        AuthSettingsView
	Email       EmailSettingsView
	Storage     StorageSettingsView
	Security    SecuritySettingsView
	Token       TokenSettingsView
}

type ServerSettingsView struct {
	Host         string
	Port         int
	BaseURL      string
	ReadTimeout  string
	WriteTimeout string
	IdleTimeout  string
}

type DatabaseSettingsView struct {
	Type         string
	Path         string
	MaxOpenConns int
	MaxIdleConns int
}

type AuthSettingsView struct {
	Method          string
	OIDCEnabled     bool
	OIDCIssuerURL   string
	OIDCClientID    string
	OIDCRedirectURL string
	ForwardAuthEnabled bool
}

type EmailSettingsView struct {
	SMTPHost           string
	SMTPPort           int
	SMTPUser           string
	SMTPPasswordSet    bool
	FromEmail          string
	FromName           string
}

type StorageSettingsView struct {
	Type       string
	LocalPath  string
	S3Bucket   string
	S3Region   string
	S3Endpoint string
}

type SecuritySettingsView struct {
	SessionDuration string
	TokenExpiry     string
	HMACKeySet      bool
}

type TokenSettingsView struct {
	HashingEnabled bool
	SecretSet      bool
}

func ConfigToSettingsView(cfg *config.Config) SettingsView {
	view := SettingsView{
		Server: ServerSettingsView{
			Host:         cfg.Server.Host,
			Port:         cfg.Server.Port,
			BaseURL:      cfg.Server.BaseURL,
			ReadTimeout:  cfg.Server.ReadTimeout.String(),
			WriteTimeout: cfg.Server.WriteTimeout.String(),
			IdleTimeout:  cfg.Server.IdleTimeout.String(),
		},
		Database: DatabaseSettingsView{
			Type:         cfg.Database.Type,
			Path:         cfg.Database.Path,
			MaxOpenConns: cfg.Database.MaxOpenConns,
			MaxIdleConns: cfg.Database.MaxIdleConns,
		},
		Auth: AuthSettingsView{
			OIDCEnabled:        cfg.OIDC.Enabled,
			OIDCIssuerURL:      cfg.OIDC.IssuerURL,
			OIDCClientID:       cfg.OIDC.ClientID,
			OIDCRedirectURL:    cfg.OIDC.RedirectURL,
			ForwardAuthEnabled: cfg.ForwardAuth.Enabled,
		},
		Email: EmailSettingsView{
			SMTPHost:        cfg.Email.SMTPHost,
			SMTPPort:        cfg.Email.SMTPPort,
			SMTPUser:        cfg.Email.SMTPUser,
			SMTPPasswordSet: cfg.Email.SMTPPassword != "",
			FromEmail:       cfg.Email.FromEmail,
			FromName:        cfg.Email.FromName,
		},
		Storage: StorageSettingsView{
			Type:       cfg.Storage.Type,
			LocalPath:  cfg.Storage.LocalPath,
			S3Bucket:   cfg.Storage.S3Bucket,
			S3Region:   cfg.Storage.S3Region,
			S3Endpoint: cfg.Storage.S3Endpoint,
		},
		Security: SecuritySettingsView{
			SessionDuration: cfg.Security.SessionDuration.String(),
			TokenExpiry:     cfg.Security.TokenExpiry.String(),
			HMACKeySet:      cfg.Security.HMACSecretKey != "",
		},
		Token: TokenSettingsView{
			HashingEnabled: cfg.Token.HashingEnabled,
			SecretSet:      cfg.Token.Secret != "",
		},
	}

	if cfg.OIDC.Enabled {
		view.Auth.Method = "OIDC"
	} else if cfg.ForwardAuth.Enabled {
		view.Auth.Method = "Forward Auth"
	} else {
		view.Auth.Method = "None"
	}

	return view
}

type SettingsHandler struct {
	config    *config.Config
	templates *template.Template
}

type SettingsPageData struct {
	ActivePage string
	IsAdmin    bool
	User       *models.User
	Settings   SettingsView
	Error      string
}

func NewSettingsHandler(cfg *config.Config) *SettingsHandler {
	return &SettingsHandler{config: cfg}
}

func (h *SettingsHandler) SetTemplates(tmpl *template.Template) {
	h.templates = tmpl
}

func (h *SettingsHandler) SettingsPage(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		HandleError(w, r, &models.PermissionDeniedError{
			Action:   "view admin settings",
			Resource: "Admin Settings",
		})
		return
	}

	view := ConfigToSettingsView(h.config)

	data := &SettingsPageData{
		ActivePage: "admin",
		IsAdmin:    isAdminRequest(r),
		User:       user,
		Settings:   view,
	}

	h.renderPage(w, http.StatusOK, data)
}

func (h *SettingsHandler) renderPage(w http.ResponseWriter, status int, data *SettingsPageData) {
	renderHTML(w, h.templates, "admin_settings.html", status, data)
}
