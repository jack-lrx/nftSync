package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gavin/nftSync/internal/dao"
	"time"
)

const (
	NFTDetailCacheTTL = 5 * time.Minute
	NFTListCacheTTL   = 2 * time.Minute
)

// NFTService 业务逻辑层
// 依赖 NFTRepository
type NFTService struct {
	Repo *dao.NFTRepository
}

// NFTItemDTO 用于安全输出 NFT 属性
type NFTItemDTO struct {
	Name      string `json:"name,omitempty"`
	TraitType string `json:"trait_type,omitempty"`
	Value     string `json:"value,omitempty"`
}

// NFTDetailDTO 用于安全输出 NFT 详情
type NFTDetailDTO struct {
	Contract string       `json:"contract"`
	TokenID  string       `json:"token_id"`
	Owner    string       `json:"owner"`
	TokenURI string       `json:"token_uri,omitempty"`
	Metadata string       `json:"metadata,omitempty"`
	Items    []NFTItemDTO `json:"items,omitempty"`
}

// ToNFTDetailDTO 转换函数
func ToNFTDetailDTO(nft *dao.NFT) *NFTDetailDTO {
	if nft == nil {
		return nil
	}
	items := make([]NFTItemDTO, 0, len(nft.Items))
	for _, item := range nft.Items {
		items = append(items, NFTItemDTO{
			Name:      item.Name,
			TraitType: item.TraitType,
			Value:     item.Value,
		})
	}
	return &NFTDetailDTO{
		Contract: nft.Contract,
		TokenID:  nft.TokenID,
		Owner:    nft.Owner,
		TokenURI: nft.TokenURI,
		Metadata: nft.Metadata,
		Items:    items,
	}
}

// ToNFTDetailDTOList 批量转换
func ToNFTDetailDTOList(nfts []dao.NFT) []NFTDetailDTO {
	res := make([]NFTDetailDTO, 0, len(nfts))
	for _, nft := range nfts {
		dto := ToNFTDetailDTO(&nft)
		if dto != nil {
			res = append(res, *dto)
		}
	}
	return res
}

// 查询 NFT 详情，优先查 Redis，未命中查 Repo 并回写缓存，直接返回 DTO
func (s *NFTService) GetNFTDetail(ctx context.Context, contract, tokenID string) (*NFTDetailDTO, error) {
	cacheKey := fmt.Sprintf("nft:detail:%s:%s", contract, tokenID)
	cacheVal, err := GetCache(ctx, cacheKey)
	if err == nil && cacheVal != "" {
		var nft dao.NFT
		if jsonErr := json.Unmarshal([]byte(cacheVal), &nft); jsonErr == nil {
			return ToNFTDetailDTO(&nft), nil
		}
	}
	// 未命中缓存，查 Repo
	nft, err := s.Repo.GetNFTDetail(contract, tokenID)
	if err != nil {
		return nil, err
	}
	// 回写缓存
	if data, jsonErr := json.Marshal(nft); jsonErr == nil {
		SetCache(ctx, cacheKey, string(data), NFTDetailCacheTTL)
	}
	return ToNFTDetailDTO(nft), nil
}

// 查询某个 owner 的所有 NFT，优先查 Redis，未命中查 Repo 并回写缓存，直接返回 DTO 列表
func (s *NFTService) GetNFTListByOwner(ctx context.Context, owner string) ([]NFTDetailDTO, error) {
	cacheKey := fmt.Sprintf("nft:list:owner:%s", owner)
	cacheVal, err := GetCache(ctx, cacheKey)
	if err == nil && cacheVal != "" {
		var nfts []dao.NFT
		if jsonErr := json.Unmarshal([]byte(cacheVal), &nfts); jsonErr == nil {
			return ToNFTDetailDTOList(nfts), nil
		}
	}
	// 未命中缓存，查 Repo
	nfts, err := s.Repo.GetNFTListByOwner(owner)
	if err != nil {
		return nil, err
	}
	// 回写缓存
	if data, jsonErr := json.Marshal(nfts); jsonErr == nil {
		SetCache(ctx, cacheKey, string(data), NFTListCacheTTL)
	}
	return ToNFTDetailDTOList(nfts), nil
}
