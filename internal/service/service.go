package service

import (
	"github.com/gavin/nftSync/internal/config"
	"github.com/gavin/nftSync/internal/dao"
	"github.com/gavin/nftSync/internal/middleware"
)

type Service struct {
	Dao   *dao.Dao
	Cache *middleware.Cache
}

func NewService(ctx *config.Context) *Service {
	bizDao := dao.New(ctx.Db)
	cache := middleware.NewRedis(ctx.Redis)

	return &Service{
		Dao:   bizDao,
		Cache: cache,
	}
}
