package service

import (
	"github.com/gavin/nftSync/internal/blockchain"
	"github.com/gavin/nftSync/internal/config"
	"github.com/gavin/nftSync/internal/dao"
	"math/big"
)

type MultiNodeSyncService struct {
	MultiNode       *blockchain.MultiNodeEthClient
	lastSyncedBlock *big.Int
	Dao             *dao.Dao
}

func NewMultiNodeSyncService(ctx *config.Context) *MultiNodeSyncService {
	return &MultiNodeSyncService{
		MultiNode:       ctx.MultiNode,
		lastSyncedBlock: big.NewInt(0),
		Dao:             dao.New(ctx.Db),
	}
}
