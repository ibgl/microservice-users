package adapters

import (
	"context"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/ibgl/microservice-users/internal/app/jwt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshModel struct {
	UUID      uuid.UUID `db:"uuid"`
	UserUUID  uuid.UUID `db:"user_uuid"`
	Token     string    `db:"token"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type RefreshPgsqlRepository struct {
	pool *pgxpool.Pool
}

func NewRefreshPgsqlRepository(pool *pgxpool.Pool) *RefreshPgsqlRepository {
	return &RefreshPgsqlRepository{pool}
}

func (s *RefreshPgsqlRepository) Add(ctx context.Context, refresh *jwt.RefreshJWT) error {
	_, err := s.pool.Exec(ctx, "insert into refresh_tokens(uuid, user_uuid, token, created_at, updated_at) values($1,$2,$3,$4,$5)",
		refresh.Claims.UUID,
		refresh.Claims.UserId,
		refresh.Token,
		time.Now(),
		time.Now())

	if err != nil {
		return err
	}

	return nil
}
func (s *RefreshPgsqlRepository) Exists(ctx context.Context, uuid, userUUID uuid.UUID, token string) (bool, error) {
	model := &RefreshModel{}
	if err := pgxscan.Get(
		ctx, s.pool, model, "select * from refresh_tokens where uuid = $1 and user_uuid = $2 and token = $3",
		uuid,
		userUUID,
		token,
	); err != nil {
		if pgxscan.NotFound(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (s *RefreshPgsqlRepository) Delete(ctx context.Context, uuid uuid.UUID) error {
	_, err := s.pool.Exec(ctx, "delete from refresh_tokens where uuid = $1", uuid)

	if err != nil {
		return err
	}

	return nil
}

func (s *RefreshPgsqlRepository) DeleteForUserUUID(ctx context.Context, userUUID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, "delete from refresh_tokens where user_uuid = $1", userUUID)

	if err != nil {
		return err
	}

	return nil
}

func (s *RefreshPgsqlRepository) CountForUser(ctx context.Context, userUUID uuid.UUID) (int, error) {
	var counter int

	err := s.pool.QueryRow(ctx, "SELECT count(*) FROM refresh_tokens where user_uuid = $1", userUUID).Scan(&counter)
	return counter, err
}
