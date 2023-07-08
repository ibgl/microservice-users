package mocks

import (
	"context"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/ibgl/microservice-users/internal/app"
	"github.com/ibgl/microservice-users/internal/app/user"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
)

// App mocked structure
type AppMock struct {
	authService user.UserAuthManager
	logger      *logrus.Logger
	config      *app.Config
}

func (h *AppMock) GetConfig() *app.Config {
	return h.config
}

func (h *AppMock) GetLogger() *logrus.Logger {
	return h.logger
}

func (h *AppMock) GetAuthService() user.UserAuthManager {
	return h.authService
}

func (h *AppMock) SetConfig(config *app.Config) app.Application {
	h.config = config
	return h
}

func (h *AppMock) SetLogger(logger *logrus.Logger) app.Application {
	h.logger = logger
	return h
}

func (h *AppMock) SetAuthService(authService user.UserAuthManager) app.Application {
	h.authService = authService
	return h
}

// Auth mocked structure
type AuthServiceMock struct {
	mock.Mock
}

func (m *AuthServiceMock) SignIn(ctx context.Context, r *user.SignInRequest) (*user.LoginResponse, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*user.LoginResponse), args.Error(1)
}

func (m *AuthServiceMock) SignUp(ctx context.Context, r *user.SignUpRequest) (*user.LoginResponse, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*user.LoginResponse), args.Error(1)
}

func (m *AuthServiceMock) GetUser(ctx context.Context, userUUID uuid.UUID) (*user.User, error) {
	args := m.Called(ctx, userUUID)
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *AuthServiceMock) SaveRefresh(ctx context.Context, userUUID uuid.UUID) (*user.User, error) {
	args := m.Called(ctx, userUUID)
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *AuthServiceMock) Refresh(ctx context.Context, tokenString string) (*user.LoginResponse, error) {
	args := m.Called(ctx, tokenString)
	return args.Get(0).(*user.LoginResponse), args.Error(1)
}

func (m *AuthServiceMock) UpdateSettings(ctx context.Context, request *user.UpdateSettingsRequest) (*user.User, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *AuthServiceMock) ValidateToken(ctx context.Context, token string) (user.Token, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(user.Token), args.Error(1)
}

func (m *AuthServiceMock) GoogleSignIn(ctx context.Context, token string) (*user.LoginResponse, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(*user.LoginResponse), args.Error(1)
}

func NewAppMock(
	config *app.Config,
) app.Application {
	loggerMock := &logrus.Logger{
		Out:   os.Stdout,
		Level: logrus.DebugLevel,
		Formatter: &logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
		},
	}

	if config == nil {
		config = getDefaultConfigMock()
	}

	return &AppMock{
		authService: &AuthServiceMock{},
		logger:      loggerMock,
		config:      config,
	}
}

func getDefaultConfigMock() *app.Config {
	JwtSecretMock := "secret"
	JwtAccessTTLMock := 10
	JwtRefreshTTLMock := 10
	MaxUsersSessionMock := 5

	return &app.Config{
		JwtSecret:       JwtSecretMock,
		JwtAccessTTL:    JwtAccessTTLMock,
		JwtRefreshTTL:   JwtRefreshTTLMock,
		MaxUserSessions: MaxUsersSessionMock,
	}
}
