package auth

import (
	"context"

	dbsqlc "github.com/alberdjuniawan/votesystem/internal/db/sqlc"
	"github.com/alberdjuniawan/votesystem/internal/shared/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	queries *dbsqlc.Queries
	db      *pgxpool.Pool
}

func NewRepository(queries *dbsqlc.Queries, db *pgxpool.Pool) *Repository {
	return &Repository{queries: queries, db: db}
}

func (r *Repository) CreateUser(ctx context.Context, email, passwordHash, name string) (dbsqlc.User, error) {
	return r.queries.CreateUser(ctx, dbsqlc.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
	})
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (dbsqlc.User, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

func (r *Repository) GetUserByID(ctx context.Context, id string) (dbsqlc.User, error) {
	uid, err := utils.StrToPgUUID(id)
	if err != nil {
		return dbsqlc.User{}, err
	}
	return r.queries.GetUserByID(ctx, uid)
}
