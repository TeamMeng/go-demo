package grpc

import "context"

type User struct {
}

type UserService interface {
	GetById(ctx context.Context, id int64) (User, error)
}

type UserServiceHttpImpl struct {
}

func (s *UserServiceHttpImpl) GetById(ctx context.Context, id int64) (User, error) {
	// TODO: implement
	return User{}, nil
}
