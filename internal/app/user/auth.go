package user

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ibgl/microservice-users/internal/app/currency"
	"github.com/ibgl/microservice-users/internal/app/day"
	appErr "github.com/ibgl/microservice-users/internal/app/errors"
	appJwt "github.com/ibgl/microservice-users/internal/app/jwt"
	"google.golang.org/api/idtoken"
)

type LoginResponse struct {
	Access  Token
	Refresh Token
}

type Token struct {
	Value  string
	UserId uuid.UUID
}

type SignInRequest struct {
	Email    string
	Password string
}

type SignUpRequest struct {
	Email    string
	Password string
	Name     string
}

type UpdateSettingsRequest struct {
	UserUUID          uuid.UUID
	Currency          string
	FirstDayOfWeek    string
	ProfilePictureUrl string
}

type AuthService struct {
	userRepository    UserRepository
	jwtService        *appJwt.JWTService
	refreshRepository RefreshJWTRepository
	maxUserSessions   int
	googleKey         string
}

type RefreshJWTRepository interface {
	Add(ctx context.Context, refresh *appJwt.RefreshJWT) error
	Exists(ctx context.Context, uuid, userUUID uuid.UUID, token string) (bool, error)
	Delete(ctx context.Context, uuid uuid.UUID) error
	DeleteForUserUUID(ctx context.Context, userUUID uuid.UUID) error
	CountForUser(ctx context.Context, userUUID uuid.UUID) (int, error)
}

func NewAuthService(ur UserRepository, jwt *appJwt.JWTService, rfr RefreshJWTRepository, mus int, googleKey string) *AuthService {
	return &AuthService{
		userRepository:    ur,
		jwtService:        jwt,
		refreshRepository: rfr,
		maxUserSessions:   mus,
		googleKey:         googleKey,
	}
}

func (h *AuthService) ValidateToken(ctx context.Context, token string) (Token, error) {
	access, err := h.jwtService.ValidateAccess(token)
	if err != nil {
		return Token{}, err
	}

	return Token{
		Value:  access.Token,
		UserId: access.Claims.UserId,
	}, nil
}

func (h *AuthService) SignIn(ctx context.Context, r *SignInRequest) (*LoginResponse, error) {
	userFound, err := h.userRepository.FindByEmail(ctx, r.Email)
	if err != nil {
		return &LoginResponse{}, appErr.NewIncorrectInputError("User not found", "invalid-credentials")
	}

	if !CheckPasswordHash(r.Password, userFound.Hash) {
		return &LoginResponse{}, appErr.NewIncorrectInputError("User not found", "invalid-credentials")
	}

	return h.createTokens(ctx, userFound)
}

func (h *AuthService) SignUp(ctx context.Context, r *SignUpRequest) (*LoginResponse, error) {
	_, err := h.userRepository.FindByEmail(ctx, r.Email)
	if err != nil {
		if !appErr.IsNotFound(err) {
			return &LoginResponse{}, err
		}
	} else {
		return &LoginResponse{}, appErr.NewIncorrectInputError("Email already in use", "field-email-invalid")
	}

	hash, err := HashPassword(r.Password)
	if err != nil {
		return &LoginResponse{}, appErr.NewAppError(err.Error(), "create-user-error")
	}

	user := NewUser(
		uuid.New(),
		r.Email,
		r.Name,
		hash,
		DefaultUserSettings(),
		time.Now(),
		time.Now(),
	)

	err = h.userRepository.Transactional(ctx, func(r UserRepository) error {
		return h.userRepository.Add(ctx, user)
	})

	if err != nil {
		return &LoginResponse{}, appErr.NewAppError(err.Error(), "user-saving-error")
	}

	return h.createTokens(ctx, user)
}

func (h *AuthService) createTokens(ctx context.Context, user *User) (*LoginResponse, error) {
	accessClaims := appJwt.NewAccessClaims(
		user.UUID,
		user.Email,
		user.Name,
	)

	refreshClaims := appJwt.NewRefreshClaims(
		user.UUID,
	)

	var access *appJwt.AccessJWT
	var refresh *appJwt.RefreshJWT
	var err error
	var wg sync.WaitGroup

	go func() {
		access, err = h.jwtService.CreateAccess(*accessClaims)
		wg.Done()
	}()

	go func() {
		refresh, err = h.jwtService.CreateRefresh(*refreshClaims)
		wg.Done()
	}()

	wg.Add(2)

	if err != nil {
		return &LoginResponse{}, appErr.NewAuthorizationError(err.Error(), "could-not-authorize-user")
	}

	refreshCount, err := h.refreshRepository.CountForUser(ctx, refreshClaims.UserId)
	if err != nil {
		return &LoginResponse{}, appErr.NewAuthorizationError(err.Error(), "could-not-authorize-user")
	}

	if refreshCount > h.maxUserSessions {
		err = h.refreshRepository.DeleteForUserUUID(ctx, refreshClaims.UserId)
		if err != nil {
			return &LoginResponse{}, appErr.NewAuthorizationError(err.Error(), "could-not-authorize-user")
		}
	}

	err = h.refreshRepository.Add(ctx, refresh)
	if err != nil {
		return &LoginResponse{}, appErr.NewAuthorizationError(err.Error(), "could-not-authorize-user")
	}

	return &LoginResponse{
		Access: Token{
			access.Token,
			access.Claims.UserId,
		},
		Refresh: Token{
			refresh.Token,
			refresh.Claims.UserId,
		},
	}, nil
}

func (h *AuthService) GetUser(ctx context.Context, userUUID uuid.UUID) (*User, error) {
	return h.userRepository.FindById(ctx, userUUID)
}

func (h *AuthService) SaveRefresh(ctx context.Context, userUUID uuid.UUID) (*User, error) {
	return h.userRepository.FindById(ctx, userUUID)
}

func (h *AuthService) Refresh(ctx context.Context, tokenString string) (*LoginResponse, error) {
	refresh, err := h.jwtService.ValidateRefresh(tokenString)
	if err != nil {
		return &LoginResponse{}, appErr.NewAuthorizationError(err.Error(), "invalid-token")
	}

	exists, err := h.refreshRepository.Exists(ctx, refresh.Claims.UUID, refresh.Claims.UserId, tokenString)
	if err != nil {
		return &LoginResponse{}, appErr.NewAuthorizationError(err.Error(), "invalid-token")
	}

	if !exists {
		return &LoginResponse{}, appErr.NewAuthorizationError("Refresh not found", "invalid-token")
	}

	err = h.refreshRepository.Delete(ctx, refresh.Claims.UUID)
	if err != nil {
		return &LoginResponse{}, appErr.NewAuthorizationError(err.Error(), "invalid-token")
	}

	user, err := h.userRepository.FindById(ctx, refresh.Claims.UserId)
	if err != nil {
		return &LoginResponse{}, appErr.NewAuthorizationError(err.Error(), "invalid-token")
	}

	return h.createTokens(ctx, user)
}

func (h *AuthService) UpdateSettings(ctx context.Context, request *UpdateSettingsRequest) (*User, error) {
	cur, err := currency.FromString(request.Currency)
	if err != nil {
		return &User{}, appErr.NewIncorrectInputError("Invalid currency", "field-currency-invalid")
	}

	day, err := day.FromString(request.FirstDayOfWeek)
	if err != nil {
		return &User{}, appErr.NewIncorrectInputError("Invalid first day of week", "field-first-day-of-week-invalid")
	}

	settings := NewUserSettings(cur, day, request.ProfilePictureUrl)

	_, err = h.userRepository.FindById(ctx, request.UserUUID)
	if err != nil {
		return &User{}, err
	}

	return h.userRepository.UpdateSettings(ctx, request.UserUUID, &settings)
}

func (h *AuthService) GoogleSignIn(ctx context.Context, tokenString string) (*LoginResponse, error) {
	validTok, err := idtoken.Validate(ctx, tokenString, h.googleKey)
	if err != nil {
		return &LoginResponse{}, appErr.NewIncorrectInputError(fmt.Sprintf("%v", err), "invalid-credentials")
	}

	authUser, err := h.userRepository.FindByEmail(ctx, fmt.Sprintf("%s", validTok.Claims["email"]))

	if err == nil {
		log.Printf("user found %v %v", authUser, err)
		settings := authUser.Settings
		settings.ProfilePictureUrl = fmt.Sprintf("%s", validTok.Claims["picture"])
		h.userRepository.UpdateSettings(ctx, authUser.UUID, &settings)

		return h.createTokens(ctx, authUser)
	} else if !appErr.IsNotFound(err) {
		return &LoginResponse{}, err
	}

	hash, err := HashPassword(GeneratePassword())
	if err != nil {
		return &LoginResponse{}, appErr.NewAppError(err.Error(), "create-user-error")
	}

	settings := DefaultUserSettings()
	settings.ProfilePictureUrl = fmt.Sprintf("%s", validTok.Claims["picture"])

	authUser = NewUser(
		uuid.New(),
		fmt.Sprintf("%s", validTok.Claims["email"]),
		fmt.Sprintf("%s", validTok.Claims["name"]),
		hash,
		DefaultUserSettings(),
		time.Now(),
		time.Now(),
	)

	err = h.userRepository.Add(ctx, authUser)
	if err != nil {
		return &LoginResponse{}, appErr.NewAppError(err.Error(), "user-saving-error")
	}

	return h.createTokens(ctx, authUser)
}
