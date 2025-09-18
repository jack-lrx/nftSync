package dao

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	Email        string    `json:"email" gorm:"uniqueIndex"`
	PasswordHash string    `json:"-"`
	WalletAddr   string    `json:"wallet_addr" gorm:"uniqueIndex"`
	CreatedAt    time.Time `json:"created_at"`
}

// UserRepository 用户数据访问对象
// 推荐在 service 层注入 DB 实例，避免全局 DB
type UserRepository struct {
	DB *gorm.DB
}

// 创建用户
func (r *UserRepository) CreateUser(user *User) error {
	return r.DB.Create(user).Error
}

// 通过邮箱查找用户
func (r *UserRepository) GetUserByEmail(email string) (*User, error) {
	var user User
	if err := r.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// 通过钱包地址查找用户
func (r *UserRepository) GetUserByWallet(walletAddr string) (*User, error) {
	var user User
	if err := r.DB.Where("wallet_addr = ?", walletAddr).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// 判断用户是否存在（支持邮箱或钱包地址）
func (r *UserRepository) UserExists(email, walletAddr string) (bool, error) {
	if email != "" {
		user, err := r.GetUserByEmail(email)
		if err != nil {
			return false, err
		}
		if user != nil {
			return true, nil
		}
	}
	if walletAddr != "" {
		user, err := r.GetUserByWallet(walletAddr)
		if err != nil {
			return false, err
		}
		if user != nil {
			return true, nil
		}
	}
	return false, nil
}
