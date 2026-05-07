package repository

import (
	"context"
	"database/sql"
	"time"

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
	return r.dao.Insert(ctx, r.domainToEntity(u))
}

func (r *UserRepository) FindUserByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindUserByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}

	return r.entityToDomain(u), nil
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

	u = r.entityToDomain(user)

	go func() {
		err = r.cache.Set(ctx, u)
		if err != nil {
			// info log
		}
	}()

	return u, nil
}

func (r *UserRepository) FindUserByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := r.dao.FindUserByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, nil
	}

	return r.entityToDomain(u), nil
}

func (r *UserRepository) domainToEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Password: u.Password,
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Ctime: u.Ctime.UnixMilli(),
	}
}

func (r *UserRepository) entityToDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Password: u.Password,
		Phone:    u.Phone.String,
		Ctime:    time.UnixMilli(u.Ctime),
	}
}
