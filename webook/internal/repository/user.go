package repository

import (
	"context"

	"github.com/TeamMeng/go-demo/webook/internal/domain"
	"github.com/TeamMeng/go-demo/webook/internal/repository/dao"
)

var (
	ErrUserDuplicalicate = dao.ErrUserDuplicalicate
	ErrUserNotFound      = dao.ErrUserNotFound
)

type UserRepository struct {
	dao *dao.UserDAO
}

func NewUserRepository(dao *dao.UserDAO) *UserRepository {
	return &UserRepository{
		dao: dao,
	}
}

func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
}

func (r *UserRepository) FindUserByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindUserByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}

	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
	}, nil
}
