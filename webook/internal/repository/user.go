package repository

import (
	"context"

	"github.com/TeamMeng/go-demo/webook/internal/domain"
	"github.com/TeamMeng/go-demo/webook/internal/repository/dao"
)

var ErrUserDuplicalicate = dao.ErrUserDuplicalicate

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
