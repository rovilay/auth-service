package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rovilay/auth-service/models"
	"github.com/rovilay/auth-service/utils"
	"github.com/rs/zerolog"
)

type postgresRepository struct {
	db  *sqlx.DB
	log *zerolog.Logger
}

func NewPostgresRepository(ctx context.Context, db *sqlx.DB, log *zerolog.Logger) *postgresRepository {
	logger := log.With().Str("repository", "postgresRepository").Logger()

	// ping db
	if err := db.PingContext(ctx); err != nil {
		logger.Fatal().Err(err).Msg("[ERROR] failed to connect to postgres")
	}

	return &postgresRepository{
		log: &logger,
		db:  db,
	}
}

func (r *postgresRepository) CreateUser(ctx context.Context, user *models.User) error {
	log := r.log.With().Str("method", "CreateUser").Logger()

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		log.Err(err).Msg(utils.ErrPasswordHash.Error())
		return utils.ErrPasswordHash
	}

	query := `
		INSERT INTO users (id, firstname, lastname, username, email, password, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`

	_, err = r.db.ExecContext(ctx, query, user.ID, user.Firstname, user.Lastname, user.Username, user.Email, hashedPassword)
	if err != nil {
		return r.mapDatabaseError(err, &log)
	}

	return nil
}

func (r *postgresRepository) GetUserByIDorEmail(ctx context.Context, idOrEmail string) (*models.User, error) {
	log := r.log.With().Str("method", "GetUserByEmail").Logger()

	query := `SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL`
	// Check if the provided string looks like a UUID
	if _, err := uuid.Parse(idOrEmail); err == nil {
		// Search by UUID
		query = `SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL`
	}

	row := r.db.QueryRow(query, idOrEmail)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, r.mapDatabaseError(err, &log)
	}

	return &user, nil
}

func (r *postgresRepository) CheckUserNameExist(ctx context.Context, username string) (bool, error) {
	log := r.log.With().Str("method", "CheckUserNameExist").Logger()

	query := "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)"

	var exists bool
	err := r.db.QueryRowContext(ctx, query, username).Scan(&exists)
	if err != nil {
		return true, r.mapDatabaseError(err, &log)
	}

	return exists, nil
}

func (r *postgresRepository) mapDatabaseError(err error, log *zerolog.Logger) error {
	log.Err(err).Msg("Database operation failed!")

	var pqErr *pgconn.PgError
	if ok := errors.As(err, &pqErr); ok {
		log.Debug().Msg(fmt.Sprintf("%v:%v", ok, pqErr.SQLState()))

		switch pqErr.Code {
		case "23505": // Unique constraint violation
			return utils.ErrDuplicateEntry
		case "23503": // Foreign key violation
			return utils.ErrForeignKeyViolation
		default:
			return fmt.Errorf("database error (%s): %w", pqErr.Code, err)
		}
	} else if errors.Is(err, sql.ErrNoRows) {
		return utils.ErrNotFound
	} else {
		return fmt.Errorf("database error: %w", err)
	}
}
