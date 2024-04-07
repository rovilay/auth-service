package repository

import (
	"context"

	"github.com/rovilay/auth-service/models"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByIDorEmail(ctx context.Context, idOrEmail string) (*models.User, error)
	CheckUserNameExist(ctx context.Context, username string) (bool, error)
}
