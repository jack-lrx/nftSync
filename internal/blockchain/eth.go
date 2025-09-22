package blockchain

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gavin/nftSync/internal/blockchain/erc721"
	"math/big"
	"os"
)

// EthClient 封装以太坊客户端
type EthClient struct {
	client *ethclient.Client
}

func NewEthClient(client *ethclient.Client) *EthClient {
	return &EthClient{client}
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

// FetchOrderFilledEvents 拉取市场合约的订单成交事件（接口/伪代码，需结合ABI实现）
func (e *EthClient) FetchOrderFilledEvents(ctx context.Context, contract string, startBlock, endBlock *big.Int) ([]OrderFilledEvent, error) {
	// 市场合约ABI路径（后续可参数化或配置）
	abiPath := "internal/blockchain/marketplace/Marketplace.abi"
	abiFile, err := os.Open(abiPath)
	if err != nil {
		return nil, err
	}
	defer abiFile.Close()
	marketAbi, err := abi.JSON(abiFile)
	if err != nil {
		return nil, err
	}

	// 事件签名（需与ABI一致）
	eventSig := "OrderFilled(address,address,uint256,uint256,uint256)"
	eventTopic := crypto.Keccak256Hash([]byte(eventSig))

	query := ethereum.FilterQuery{
		FromBlock: startBlock,
		ToBlock:   endBlock,
		Addresses: []common.Address{common.HexToAddress(contract)},
		Topics:    [][]common.Hash{{eventTopic}},
	}
	logs, err := e.client.FilterLogs(ctx, query)
	if err != nil {
		return nil, err
	}

	var events []OrderFilledEvent
	for _, vLog := range logs {
		// 解析 topics
		if len(vLog.Topics) < 4 {
			continue // topics数量不足
		}
		// seller, buyer, tokenId
		seller := common.HexToAddress(vLog.Topics[1].Hex()).Hex()
		buyer := common.HexToAddress(vLog.Topics[2].Hex()).Hex()
		tokenId := vLog.Topics[3].Hex()

		// 解析 data（price, fee）
		var event struct {
			Price *big.Int
			Fee   *big.Int
		}
		err := marketAbi.UnpackIntoInterface(&event, "OrderFilled", vLog.Data)
		if err != nil {
			continue // 解码失败跳过
		}

		block, err := e.client.BlockByNumber(ctx, big.NewInt(int64(vLog.BlockNumber)))
		var blockTime int64
		if err == nil {
			blockTime = int64(block.Time())
		}

		priceFloat, _ := new(big.Float).SetInt(event.Price).Float64()
		feeFloat, _ := new(big.Float).SetInt(event.Fee).Float64()
		orderEvent := OrderFilledEvent{
			TokenID:     tokenId,
			Seller:      seller,
			Buyer:       buyer,
			Price:       priceFloat,
			Fee:         feeFloat,
			Contract:    vLog.Address.Hex(),
			BlockNumber: vLog.BlockNumber,
			TxHash:      vLog.TxHash.Hex(),
			BlockTime:   blockTime,
		}
		events = append(events, orderEvent)
	}
	return events, nil
}

// TODO: 添加事件监听与合约交互方法
