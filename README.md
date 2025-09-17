# NFTSync

本项目为基于Go语言的Web3 NFT数据同步服务基础骨架，包含以下核心模块：

- 区块链节点连接（以太坊）
- NFT合约事件监听（待实现）
- 数据存储模型（GORM）
- 元数据抓取（待实现）
- API服务（Gin）

## 目录结构

```
cmd/nftSync/main.go         # 主入口，启动API服务
internal/blockchain/eth.go # 以太坊节点连接与区块获取
internal/model/nft.go      # NFT数据模型定义
internal/service/sync.go   # NFT同步服务逻辑
internal/api/server.go     # API接口服务
configs/config.yaml        # 配置文件示例
```

## 依赖安装

请在项目根目录执行：

```
go mod init github.com/gavin/nftSync

go get github.com/ethereum/go-ethereum github.com/gin-gonic/gin gorm.io/gorm gorm.io/driver/postgres
```

## 启动服务

```
go run cmd/nftSync/main.go
```

## 后续开发建议
- 实现区块链事件监听与NFT数据同步逻辑
- 集成数据库持久化
- 完善API接口（如NFT查询、元数据解析等）
- 支持多链与多合约

如需帮助可随时补充需求。

