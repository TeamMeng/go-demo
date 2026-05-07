package repository

import (
	"context"

	"github.com/TeamMeng/go-demo/webook/internal/domain"
	"github.com/TeamMeng/go-demo/webook/internal/repository/cache"
	"github.com/TeamMeng/go-demo/webook/internal/repository/dao"
)

var (
	ErrUserDuplicalicate = dao.ErrUserDuplicalicate
	ErrUserNotFound      = dao.ErrUserNotFound
)

type UserRepository struct {
	dao   *dao.UserDAO
	cache *cache.UserCache
}

func NewUserRepository(dao *dao.UserDAO, cache *cache.UserCache) *UserRepository {
	return &UserRepository{
		dao:   dao,
		cache: cache,
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

func (r *UserRepository) FindUserById(ctx context.Context, id int64) (domain.User, error) {
	u, err := r.cache.Get(ctx, id)
	if err == nil {
		return u, err
	}

	user, err := r.dao.FindUserById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	u = domain.User{
		Id:       user.Id,
		Email:    user.Email,
		Password: user.Password,
	}

	go func() {
		err = r.cache.Set(ctx, u)
		if err != nil {
			// info log
		}
	}()

	return u, nil
}
