package dao

import (
	"time"
)

type User struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	Email        string    `json:"email" gorm:"uniqueIndex"`
	PasswordHash string    `json:"-"`
	WalletAddr   string    `json:"wallet_addr" gorm:"uniqueIndex"`
	CreatedAt    time.Time `json:"created_at"`
}

// 创建用户
func CreateUser(user *User) error {
	// TODO: 实现数据库插入逻辑
	return nil
}

// 通过邮箱查找用户
func GetUserByEmail(email string) (*User, error) {
	// TODO: 实现数据库查询逻辑
	return nil, nil
}

// 通过钱包地址查找用户
func GetUserByWallet(walletAddr string) (*User, error) {
	// TODO: 实现数据库查询逻辑
	return nil, nil
}
