package adapters

import (
	"context"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/ibgl/microservice-users/internal/app/currency"
	"github.com/ibgl/microservice-users/internal/app/day"
	"github.com/ibgl/microservice-users/internal/app/errors"
	"github.com/ibgl/microservice-users/internal/app/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserModel struct {
	UUID      uuid.UUID         `db:"uuid"`
	Email     string            `db:"email"`
	Name      string            `db:"name"`
	Hash      string            `db:"hash"`
	Settings  map[string]string `db:"settings"`
	CreatedAt time.Time         `db:"created_at"`
	UpdatedAt time.Time         `db:"updated_at"`
}

type PgxConnector interface {
	pgxscan.Querier
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type UserPgsqlRepository struct {
	pool *pgxpool.Pool
	tx   pgx.Tx
}

func UserSettingsToMap(s user.UserSettings) map[string]string {
	return map[string]string{
		"currency":            s.Currency.String(),
		"first_day_of_week":   s.FirstDayOfWeek.String(),
		"profile_picture_url": s.ProfilePictureUrl,
	}
}

func NewUserPgsqlRepository(pool *pgxpool.Pool) *UserPgsqlRepository {
	return &UserPgsqlRepository{pool, nil}
}

func (s *UserPgsqlRepository) connector() PgxConnector {
	if s.tx != nil {
		return s.tx
	}

	return s.pool
}

func (s *UserPgsqlRepository) exec(ctx context.Context, sql string, arguments ...any) error {
	_, err := s.connector().Exec(ctx, sql, arguments...)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserPgsqlRepository) get(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	if err := pgxscan.Get(
		ctx, s.connector(), dst, query, args...,
	); err != nil {
		if pgxscan.NotFound(err) {
			return errors.NewNotFoundError("Not found", "not-found")
		}

		return err
	}

	return nil
}

func (s *UserPgsqlRepository) Transactional(ctx context.Context, cb func(r user.UserRepository) error) error {
	if s.tx == nil {
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return err
		}

		s.tx = tx
	}

	err := cb(s)
	if err != nil {
		s.tx.Rollback(ctx)
		s.tx = nil
		return err
	}

	err = s.tx.Commit(ctx)
	if err != nil {
		s.tx = nil
		return err
	}

	s.tx = nil
	return nil
}

func (s *UserPgsqlRepository) FindById(ctx context.Context, uuid uuid.UUID) (*user.User, error) {
	userModel := &UserModel{}
	if err := s.get(
		ctx, userModel, "select uuid, email, name, hash, settings, created_at, updated_at from users where uuid = $1", uuid,
	); err != nil {
		if err.Error() == "not-found" {
			return &user.User{}, errors.NewNotFoundError("User not found", "user-not-found")
		}

		return &user.User{}, err
	}

	return serviceUserFromModel(userModel)
}

func (s *UserPgsqlRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	userModel := &UserModel{}
	if err := s.get(
		ctx, userModel, "select uuid, email, name, hash, settings, created_at, updated_at from users where email = $1", email,
	); err != nil {
		if err.Error() == "not-found" {
			return &user.User{}, errors.NewNotFoundError("User not found", "user-not-found")
		}

		return &user.User{}, err
	}

	return serviceUserFromModel(userModel)
}

func (s *UserPgsqlRepository) Add(ctx context.Context, u *user.User) error {
	err := s.exec(ctx, "insert into users(uuid, email, name, hash, settings, created_at, updated_at) values($1,$2,$3,$4,$5,$6, $7)", u.UUID, u.Email, u.Name, u.Hash, UserSettingsToMap(u.Settings), u.CreatedAt, u.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserPgsqlRepository) UpdateSettings(ctx context.Context, user_uuid uuid.UUID, settings *user.UserSettings) (*user.User, error) {
	err := s.exec(ctx, "update users set settings = $1 where uuid = $2", UserSettingsToMap(*settings), user_uuid)
	if err != nil {
		return &user.User{}, err
	}

	return s.FindById(ctx, user_uuid)
}

func serviceUserFromModel(model *UserModel) (*user.User, error) {
	cur, err := currency.FromString(model.Settings["currency"])
	if err != nil {
		return &user.User{}, err
	}

	FirstDayOfWeek, err := day.FromString(model.Settings["first_day_of_week"])
	if err != nil {
		FirstDayOfWeek = day.MON
	}

	return user.NewUser(
		model.UUID,
		model.Email,
		model.Name,
		model.Hash,
		user.NewUserSettings(cur, FirstDayOfWeek, model.Settings["profile_picture_url"]),
		model.CreatedAt,
		model.UpdatedAt,
	), nil
}
