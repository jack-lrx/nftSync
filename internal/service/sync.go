package service

import (
	"context"
	"encoding/json"
	"github.com/gavin/nftSync/internal/blockchain"
	"github.com/gavin/nftSync/internal/config"
	"github.com/gavin/nftSync/internal/model"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"
)

// SyncService 用于同步NFT数据
type SyncService struct {
	eth             *blockchain.EthClient
	lastSyncedBlock *big.Int // 记录已同步的区块高度
	// db *gorm.DB // 可扩展数据库连接
}

func NewSyncService(eth *blockchain.EthClient) *SyncService {
	return &SyncService{
		eth:             eth,
		lastSyncedBlock: big.NewInt(0), // 初始化为0，可持久化
	}
}

// 多节点同步服务结构体
type MultiNodeSyncService struct {
	MultiNode       *blockchain.MultiNodeEthClient
	lastSyncedBlock *big.Int
	db              *gorm.DB // 可扩展数据库连接
}

func NewMultiNodeSyncService(multiNode *blockchain.MultiNodeEthClient) *MultiNodeSyncService {
	return &MultiNodeSyncService{
		MultiNode:       multiNode,
		lastSyncedBlock: big.NewInt(0),
	}
}

// 实时监听铸造事件（Transfer from=0x0）
func (s *MultiNodeSyncService) SyncMintEventsRealtime(ctx context.Context) {
	latestBlock := getLatestBlock(s.MultiNode, ctx)
	if latestBlock == nil {
		log.Printf("无法获取最新区块")
		return
	}
	startBlock := new(big.Int).Set(latestBlock)
	// 合约地址直接用全局配置
	nftContracts := config.GlobalConfig.NFTContracts
	for _, contract := range nftContracts {
		multiEvents := s.MultiNode.FetchTransferEventsAllNodes(contract, startBlock, startBlock, ctx)
		for _, mevt := range multiEvents {
			if !isMintEvent(mevt.Event) {
				continue
			}
			processMintEvent(mevt, contract, s, ctx)
		}
	}
}

// 定时轮询补全铸造事件（区块范围轮询）
func (s *MultiNodeSyncService) SyncMintEventsPolling(ctx context.Context) {
	latestBlock := getLatestBlock(s.MultiNode, ctx)
	if latestBlock == nil {
		log.Printf("无法获取最新区块")
		return
	}
	startBlock := new(big.Int).Add(s.lastSyncedBlock, big.NewInt(1))
	confirmBlocks := config.GlobalConfig.Sync.ConfirmBlocks
	safeBlock := new(big.Int).Sub(latestBlock, big.NewInt(int64(confirmBlocks)))
	if safeBlock.Cmp(startBlock) < 0 {
		log.Println("无新区块达到安全确认高度，无需轮询")
		return
	}
	nftContracts := config.GlobalConfig.NFTContracts
	for _, contract := range nftContracts {
		multiEvents := s.MultiNode.FetchTransferEventsAllNodes(contract, startBlock, safeBlock, ctx)
		for _, mevt := range multiEvents {
			if !isMintEvent(mevt.Event) {
				continue
			}
			processMintEvent(mevt, contract, s, ctx)
		}
	}
	s.lastSyncedBlock.Set(safeBlock)
	log.Printf("轮询补全完成，已安全同步到区块 %v", safeBlock)
}

// 判断是否为铸造事件（Transfer from=0x0）
func isMintEvent(evt blockchain.TransferEvent) bool {
	return evt.From == "0x0000000000000000000000000000000000000000"
}

// 获取最新区块（多节点交叉验证，取最小值，保证所有节点都能采集到数据）
func getLatestBlock(multiNode *blockchain.MultiNodeEthClient, ctx context.Context) *big.Int {
	var minBlock *big.Int
	for _, cli := range multiNode.Clients {
		blk, err := cli.GetBlockNumber(ctx)
		if err == nil {
			if minBlock == nil || blk.Cmp(minBlock) < 0 {
				minBlock = new(big.Int).Set(blk)
			}
		}
	}
	return minBlock
}

// 处理铸造事件，交叉验证、分叉检测、持久化
func processMintEvent(mevt blockchain.MultiNodeTransferEvent, contract string, s *MultiNodeSyncService, ctx context.Context) {
	confirmed := mevt.Confidence >= len(s.MultiNode.Clients)
	tokenIdBig, err := strconv.ParseInt(mevt.Event.TokenID, 0, 64)
	if err != nil {
		log.Printf("tokenId解析失败: %v", err)
		return
	}
	tokenURI, err := s.MultiNode.Clients[0].GetTokenURI(ctx, contract, big.NewInt(tokenIdBig))
	if err != nil {
		log.Printf("tokenURI获取失败: %v", err)
		return
	}
	meta, err := fetchMetadata(tokenURI)
	if err != nil {
		log.Printf("元数据获取失败: %v", err)
		return
	}
	items := []model.Item{}
	for _, attr := range meta.Attributes {
		items = append(items, model.Item{
			Name:      meta.Name,
			TraitType: attr.TraitType,
			Value:     attr.Value,
		})
	}
	metaJson, _ := json.Marshal(meta)
	nft := model.NFT{
		TokenID:     mevt.Event.TokenID,
		Contract:    mevt.Event.Contract,
		Owner:       mevt.Event.To,
		TokenURI:    tokenURI,
		Metadata:    string(metaJson),
		Price:       "", // 保持原逻辑
		Items:       items,
		Confidence:  mevt.Confidence,
		Confirmed:   confirmed,
		SourceNodes: strings.Join(mevt.SourceNodes, ","),
	}
	log.Printf("铸造NFT: %+v, Items: %+v", nft, nft.Items)
	if s.db != nil {
		repo := &model.NFTRepository{DB: s.db}
		err := s.db.Transaction(func(tx *gorm.DB) error {
			return repo.SaveOrUpdateNFT(tx, &nft)
		})
		if err != nil {
			log.Printf("NFT保存失败: %v", err)
		} else {
			log.Printf("NFT已保存或更新: tokenID=%s, contract=%s", nft.TokenID, nft.Contract)
		}
	}
}

// Attribute 表示 NFT 元数据中的单个属性
// 例如：{"trait_type": "Color", "value": "Red"}
type Attribute struct {
	TraitType string `json:"trait_type"` // 属性类型，如“Color”
	Value     string `json:"value"`      // 属性值，如“Red”
}

// Metadata 表示 NFT 的完整元数据结构
// 例如：{"name": "CryptoKitty", "description": "A cute kitty.", "image": "https://...", "attributes": [...]}
type Metadata struct {
	Name        string      `json:"name"`        // NFT名称
	Description string      `json:"description"` // NFT描述
	Image       string      `json:"image"`       // 图片URL
	Attributes  []Attribute `json:"attributes"`  // 属性列表
}

// fetchMetadata 通过tokenURI获取并解析元数据
func fetchMetadata(tokenURI string) (*Metadata, error) {
	resp, err := http.Get(tokenURI)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var meta Metadata
	if err := json.Unmarshal(body, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}
