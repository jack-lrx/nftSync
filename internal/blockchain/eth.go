package blockchain

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gavin/nftSync/internal/blockchain/erc721"
	"math/big"
)

// EthClient 封装以太坊客户端
type EthClient struct {
	client *ethclient.Client
}

func NewEthClient(client *ethclient.Client) *EthClient {
	return &EthClient{client}
}

func (e *EthClient) GetBlockByNumber(ctx context.Context, number uint64) (*types.Block, error) {
	return e.client.BlockByNumber(ctx, big.NewInt(int64(number)))
}

func (e *EthClient) GetBlockNumber(ctx context.Context) (*big.Int, error) {
	num, err := e.client.BlockNumber(ctx)
	if err != nil {
		return nil, err
	}
	return big.NewInt(int64(num)), nil
}

// TransferEvent 结构体
type TransferEvent struct {
	TokenID     string
	From        string
	To          string
	Contract    string
	BlockNumber uint64
	TxHash      string
	BlockTime   int64
}

// ERC721 Transfer事件的topic
var transferEventTopic = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)")).Hex()

// FetchTransferEvents 拉取指定区块范围内的ERC721 Transfer事件
func (e *EthClient) FetchTransferEvents(ctx context.Context, contract string, startBlock, endBlock *big.Int) ([]TransferEvent, error) {
	query := ethereum.FilterQuery{
		FromBlock: startBlock,
		ToBlock:   endBlock,
		Addresses: []common.Address{common.HexToAddress(contract)},
		Topics:    [][]common.Hash{{common.HexToHash(transferEventTopic)}},
	}
	logs, err := e.client.FilterLogs(ctx, query)
	if err != nil {
		return nil, err
	}
	var events []TransferEvent
	for _, vLog := range logs {
		if len(vLog.Topics) != 4 {
			continue // ERC721 Transfer事件应有4个topic
		}
		block, err := e.client.BlockByNumber(ctx, big.NewInt(int64(vLog.BlockNumber)))
		var blockTime int64
		if err == nil {
			blockTime = int64(block.Time())
		}
		event := TransferEvent{
			From:        common.HexToAddress(vLog.Topics[1].Hex()).Hex(),
			To:          common.HexToAddress(vLog.Topics[2].Hex()).Hex(),
			TokenID:     vLog.Topics[3].Hex(),
			Contract:    vLog.Address.Hex(),
			BlockNumber: vLog.BlockNumber,
			TxHash:      vLog.TxHash.Hex(),
			BlockTime:   blockTime,
		}
		events = append(events, event)
	}
	return events, nil
}

// GetTokenURI 通过abigen合约对象获取tokenURI
func (e *EthClient) GetTokenURI(ctx context.Context, contract string, tokenId *big.Int) (string, error) {
	address := common.HexToAddress(contract)
	instance, err := erc721.NewErc721(address, e.client)
	if err != nil {
		return "", err
	}
	uri, err := instance.TokenURI(&bind.CallOpts{Context: ctx}, tokenId)
	if err != nil {
		return "", err
	}
	return uri, nil
}

type OrderFilledEvent struct {
	TokenID     string
	Seller      string
	Buyer       string
	Price       float64
	Fee         float64
	Contract    string
	BlockNumber uint64
	TxHash      string
	BlockTime   int64
}

// FetchOrderEvents 拉取市场合约的订单相关事件（支持多事件topic参数透传）
func (e *EthClient) FetchOrderEvents(ctx context.Context, contract string, startBlock, endBlock *big.Int, topics []common.Hash) ([]types.Log, error) {
	query := ethereum.FilterQuery{
		FromBlock: startBlock,
		ToBlock:   endBlock,
		Addresses: []common.Address{common.HexToAddress(contract)},
		Topics:    [][]common.Hash{topics},
	}
	logs, err := e.client.FilterLogs(ctx, query)
	if err != nil {
		return nil, err
	}
	return logs, nil
}

// TODO: 添加事件监听与合约交互方法
