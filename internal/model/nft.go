package model

import "gorm.io/gorm"

// Item 结构体定义
type Item struct {
	ID        uint           `gorm:"primaryKey"`
	NFTID     uint           `gorm:"index;not null"` // 外键关联NFT
	Name      string         `gorm:"type:varchar(128)"`
	TraitType string         `gorm:"type:varchar(64)"`
	Value     string         `gorm:"type:varchar(128)"`
	CreatedAt int64          `gorm:"autoCreateTime:milli"`
	UpdatedAt int64          `gorm:"autoUpdateTime:milli"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// NFT 结构体定义
type NFT struct {
	ID          uint           `gorm:"primaryKey"`
	TokenID     string         `gorm:"index;not null"`
	Contract    string         `gorm:"index;not null"`
	Owner       string         `gorm:"index;not null"`
	TokenURI    string         `gorm:"type:text"`
	Metadata    string         `gorm:"type:json"`
	Price       string         `gorm:"type:varchar(64)"`
	Items       []Item         `gorm:"foreignKey:NFTID"`
	CreatedAt   int64          `gorm:"autoCreateTime:milli"`
	UpdatedAt   int64          `gorm:"autoUpdateTime:milli"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	Confidence  int            `gorm:"default:1"`     // 置信度（采集到该事件的节点数）
	Confirmed   bool           `gorm:"default:false"` // 是否已确认
	SourceNodes string         `gorm:"type:text"`     // 来源节点（逗号分隔）
}

// NFTRepository 封装 NFT 数据库操作
// SaveOrUpdateNFT 支持查重、更新、插入、属性同步，所有操作在事务中完成

// NFTRepository 用于分层管理 NFT 持久化逻辑
// 使用方法：repo := &NFTRepository{DB: db}
// repo.SaveOrUpdateNFT(tx, &nft)
type NFTRepository struct {
	DB *gorm.DB
}

// SaveOrUpdateNFT 保存或更新 NFT（带属性），所有操作在事务中完成
func (r *NFTRepository) SaveOrUpdateNFT(tx *gorm.DB, nft *NFT) error {
	var oldNFT NFT
	result := tx.Where("token_id = ? AND contract = ?", nft.TokenID, nft.Contract).First(&oldNFT)
	if result.Error == nil {
		// 已存在，更新主表和属性
		nft.ID = oldNFT.ID
		if err := tx.Model(&oldNFT).Updates(map[string]interface{}{
			"owner":        nft.Owner,
			"token_uri":    nft.TokenURI,
			"metadata":     nft.Metadata,
			"confidence":   nft.Confidence,
			"confirmed":    nft.Confirmed,
			"source_nodes": nft.SourceNodes,
		}).Error; err != nil {
			return err
		}
		// 删除旧属性
		if err := tx.Where("nft_id = ?", oldNFT.ID).Delete(&Item{}).Error; err != nil {
			return err
		}
		// 插入新属性
		for i := range nft.Items {
			nft.Items[i].NFTID = oldNFT.ID
		}
		if len(nft.Items) > 0 {
			if err := tx.Create(&nft.Items).Error; err != nil {
				return err
			}
		}
	} else if result.Error == gorm.ErrRecordNotFound {
		// 不存在，插入主表和属性
		if err := tx.Create(nft).Error; err != nil {
			return err
		}
	} else {
		return result.Error
	}
	return nil
}
