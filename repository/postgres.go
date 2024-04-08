package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
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
		RETURNING id, firstname, lastname, username, email, password, created_at, updated_at
	`

	err = r.db.QueryRowContext(
		ctx, query, user.ID, user.Firstname, user.Lastname,
		user.Username, user.Email, hashedPassword,
	).Scan(
		&user.ID, &user.Firstname, &user.Lastname,
		&user.Username, &user.Email, &user.Password,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return r.mapDatabaseError(err, &log)
	}

	return nil
}

func (r *postgresRepository) UpdateUser(ctx context.Context, user *models.User) error {
	log := r.log.With().Str("method", "UpdateUser").Logger()

	tx, err := r.db.Begin()
	if err != nil {
		return r.mapDatabaseError(err, &log)
	}
	defer tx.Rollback()

	query := `
		UPDATE users
		SET firstname = $1, lastname = $2, username = $3, email = $4, updated_at = NOW()
		WHERE id = $5 AND deleted_at IS NULL
		RETURNING id, firstname, lastname, username, email, password, created_at, updated_at
	`
	err = tx.
		QueryRowContext(ctx, query, user.Firstname, user.Lastname, user.Username, user.Email, user.ID.String()).
		Scan(
			&user.ID, &user.Firstname, &user.Lastname,
			&user.Username, &user.Email, &user.Password,
			&user.CreatedAt, &user.UpdatedAt,
		)
	if err != nil {
		return r.mapDatabaseError(err, &log)
	}

	err = tx.Commit()
	if err != nil {
		return r.mapDatabaseError(err, &log)
	}

	return nil
}

func (r *postgresRepository) UpdatePassword(ctx context.Context, userId string, password string) (*models.User, error) {
	log := r.log.With().Str("method", "UpdatePassword").Logger()

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		log.Err(err).Msg(utils.ErrPasswordHash.Error())
		return nil, utils.ErrPasswordHash
	}

	var user models.User
	tx, err := r.db.Begin()
	if err != nil {
		return nil, r.mapDatabaseError(err, &log)
	}
	defer tx.Rollback()

	query := `
		UPDATE users
		SET password = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
		RETURNING id, firstname, lastname, username, email, password, created_at, updated_at
	`
	err = tx.
		QueryRowContext(ctx, query, hashedPassword, userId).
		Scan(
			&user.ID, &user.Firstname, &user.Lastname,
			&user.Username, &user.Email, &user.Password,
			&user.CreatedAt, &user.UpdatedAt,
		)
	if err != nil {
		return nil, r.mapDatabaseError(err, &log)
	}

	err = tx.Commit()
	if err != nil {
		return nil, r.mapDatabaseError(err, &log)
	}

	return &user, nil
}

func (r *postgresRepository) GetUserByIDorEmail(ctx context.Context, idOrEmail string) (*models.User, error) {
	log := r.log.With().Str("method", "GetUserByIDorEmail").Logger()

	query := `SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL`
	// Check if the provided string looks like a UUID
	if _, err := uuid.Parse(idOrEmail); err == nil {
		// Search by UUID
		query = `SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL`
	}

	var user models.User
	err := r.db.GetContext(ctx, &user, query, idOrEmail)
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
	log.Err(err).Msg("database operation failed!")

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
