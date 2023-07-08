package app

import (
	"github.com/ibgl/microservice-users/internal/adapters"
	"github.com/ibgl/microservice-users/internal/app/jwt"
	"github.com/ibgl/microservice-users/internal/app/user"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type App struct {
	authService user.UserAuthManager
	logger      *logrus.Logger
	config      *Config
}

type Application interface {
	GetConfig() *Config
	GetLogger() *logrus.Logger
	GetAuthService() user.UserAuthManager
	SetConfig(config *Config) Application
	SetLogger(logger *logrus.Logger) Application
	SetAuthService(authService user.UserAuthManager) Application
}

func (h *App) GetConfig() *Config {
	return h.config
}

func (h *App) GetLogger() *logrus.Logger {
	return h.logger
}

func (h *App) GetAuthService() user.UserAuthManager {
	return h.authService
}

func (h *App) SetConfig(config *Config) Application {
	h.config = config
	return h
}

func (h *App) SetLogger(logger *logrus.Logger) Application {
	h.logger = logger
	return h
}

func (h *App) SetAuthService(authService user.UserAuthManager) Application {
	h.authService = authService
	return h
}

type Config struct {
	JwtSecret       string
	JwtAccessTTL    int
	JwtRefreshTTL   int
	MaxUserSessions int
	GoogleKey       string
}

func NewApplication(
	config *Config,
	logger *logrus.Logger,
	dbPool *pgxpool.Pool,
) (Application, error) {
	app := &App{}
	jwtService := jwt.NewJwtService(&jwt.JWTConfig{
		Secret:     config.JwtSecret,
		AccessTTL:  config.JwtAccessTTL,
		RefreshTTL: config.JwtRefreshTTL,
	})

	authService := user.NewAuthService(
		adapters.NewUserPgsqlRepository(dbPool),
		jwtService,
		adapters.NewRefreshPgsqlRepository(dbPool),
		config.MaxUserSessions,
		config.GoogleKey,
	)

	return app.SetAuthService(authService).SetConfig(config).SetLogger(logger), nil
}
