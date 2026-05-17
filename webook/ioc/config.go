package ioc

import "context"

type Configure interface {
	GetString(ctx context.Context, key string) (string, error)
	MustGetString(ctx context.Context, key string) string
	GetStringOrDefault(ctx context.Context, key string, def string) string
}

type myConfigure struct {
}

func (m *myConfigure) GetString(ctx context.Context, key string) (string, error) {
	panic("")
}

func (m *myConfigure) MustGetString(ctx context.Context, key string) (string, error) {
	str, err := m.GetString(ctx, key)
	if err != nil {
		panic(err)
	}

	return str, nil
}

func (m *myConfigure) GetStringOrDefault(ctx context.Context, key string, def string) string {
	str, err := m.GetString(ctx, key)
	if err != nil {
		return def
	}

	return str
}
