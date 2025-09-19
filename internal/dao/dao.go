package dao

import "gorm.io/gorm"

// Dao 订单数据访问对象
type Dao struct {
	DB *gorm.DB
}

func New(db *gorm.DB) *Dao {
	return &Dao{
		DB: db,
	}
}
