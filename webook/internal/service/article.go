package service

import (
	"context"

	"github.com/TeamMeng/go-demo/webook/internal/domain"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
}

type articleService struct{}

func NewArticleService() ArticleService {
	return &articleService{}
}

func (s *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	// TODO: implement article persistence (DAO + repository layer)
	return 0, nil
}
