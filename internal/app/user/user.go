package user

import (
	"context"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/ibgl/microservice-users/internal/app/currency"
	"github.com/ibgl/microservice-users/internal/app/day"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UUID      uuid.UUID
	Email     string
	Name      string
	Hash      string
	Settings  UserSettings
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserSettings struct {
	Currency          currency.Currency
	ProfilePictureUrl string
	FirstDayOfWeek    day.Day
}

func DefaultUserSettings() UserSettings {
	return UserSettings{
		Currency:       currency.RUB,
		FirstDayOfWeek: day.MON,
	}
}

func NewUserSettings(cur currency.Currency, firstDayOfWeek day.Day, profilePictureUrl string) UserSettings {
	return UserSettings{
		Currency:          cur,
		FirstDayOfWeek:    firstDayOfWeek,
		ProfilePictureUrl: profilePictureUrl,
	}
}

func NewUser(
	UUID uuid.UUID,
	Email string,
	Name string,
	Hash string,
	Settings UserSettings,
	CreatedAt time.Time,
	UpdatedAt time.Time,
) *User {
	return &User{
		UUID:      UUID,
		Email:     Email,
		Name:      Name,
		Hash:      Hash,
		Settings:  Settings,
		CreatedAt: CreatedAt,
		UpdatedAt: UpdatedAt,
	}
}

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" + "$#@&^*()"

type UserAuthManager interface {
	SignIn(ctx context.Context, r *SignInRequest) (*LoginResponse, error)
	SignUp(ctx context.Context, r *SignUpRequest) (*LoginResponse, error)
	GetUser(ctx context.Context, userUUID uuid.UUID) (*User, error)
	SaveRefresh(ctx context.Context, userUUID uuid.UUID) (*User, error)
	Refresh(ctx context.Context, tokenString string) (*LoginResponse, error)
	UpdateSettings(ctx context.Context, request *UpdateSettingsRequest) (*User, error)
	ValidateToken(ctx context.Context, token string) (Token, error)
	GoogleSignIn(ctx context.Context, token string) (*LoginResponse, error)
}

type UserRepository interface {
	FindById(ctx context.Context, uuid uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Add(ctx context.Context, user *User) error
	UpdateSettings(ctx context.Context, userUUID uuid.UUID, settings *UserSettings) (*User, error)
	Transactional(ctx context.Context, cb func(r UserRepository) error) error
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GeneratePassword() string {
	return StringWithCharset(20, charset)
}

func StringWithCharset(length int, charset string) string {
	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func String(length int) string {
	return StringWithCharset(length, charset)
}
