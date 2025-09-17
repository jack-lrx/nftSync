package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gavin/nftSync/internal/model"
	"gorm.io/gorm"
	"time"
)

const (
	NFTDetailCacheTTL = 5 * time.Minute
	NFTListCacheTTL   = 2 * time.Minute
)

// NFTService 业务逻辑层
type NFTService struct {
	DB *gorm.DB
}

// 查询 NFT 详情，优先查 Redis，未命中查 DB 并回写缓存
func (s *NFTService) GetNFTDetail(ctx context.Context, contract, tokenID string) (*model.NFT, error) {
	cacheKey := fmt.Sprintf("nft:detail:%s:%s", contract, tokenID)
	cacheVal, err := GetCache(ctx, cacheKey)
	if err == nil && cacheVal != "" {
		var nft model.NFT
		if jsonErr := json.Unmarshal([]byte(cacheVal), &nft); jsonErr == nil {
			return &nft, nil
		}
	}
	// 未命中缓存，查数据库
	var nft model.NFT
	err = s.DB.Preload("Items").Where("contract = ? AND token_id = ?", contract, tokenID).First(&nft).Error
	if err != nil {
		return nil, err
	}
	// 回写缓存
	if data, jsonErr := json.Marshal(nft); jsonErr == nil {
		SetCache(ctx, cacheKey, string(data), NFTDetailCacheTTL)
	}
	return &nft, nil
}

// 查询某个 owner 的所有 NFT，优先查 Redis，未命中查 DB 并回写缓存
func (s *NFTService) GetNFTListByOwner(ctx context.Context, owner string) ([]model.NFT, error) {
	cacheKey := fmt.Sprintf("nft:list:owner:%s", owner)
	cacheVal, err := GetCache(ctx, cacheKey)
	if err == nil && cacheVal != "" {
		var nfts []model.NFT
		if jsonErr := json.Unmarshal([]byte(cacheVal), &nfts); jsonErr == nil {
			return nfts, nil
		}
	}
	// 未命中缓存，查数据库
	var nfts []model.NFT
	err = s.DB.Preload("Items").Where("owner = ?", owner).Find(&nfts).Error
	if err != nil {
		return nil, err
	}
	// 回写缓存
	if data, jsonErr := json.Marshal(nfts); jsonErr == nil {
		SetCache(ctx, cacheKey, string(data), NFTListCacheTTL)
	}
	return nfts, nil
}
