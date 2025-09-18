package service

import (
	"errors"
	"github.com/gavin/nftSync/internal/dao"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	Repo *dao.UserRepository
}

// 用户注册
func (s *UserService) RegisterUser(email, password, walletAddr string) error {
	if email == "" || password == "" {
		return errors.New("邮箱和密码不能为空")
	}
	if user, _ := s.Repo.GetUserByEmail(email); user != nil {
		return errors.New("邮箱已注册")
	}
	if walletAddr != "" {
		if user, _ := s.Repo.GetUserByWallet(walletAddr); user != nil {
			return errors.New("钱包地址已注册")
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}
	user := &dao.User{
		Email:        email,
		PasswordHash: string(hash),
		WalletAddr:   walletAddr,
	}
	return s.Repo.CreateUser(user)
}

// 用户登录
func (s *UserService) LoginUser(email, password string) (*dao.User, error) {
	user, err := s.Repo.GetUserByEmail(email)
	if err != nil || user == nil {
		return nil, errors.New("用户不存在")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("密码错误")
	}
	return user, nil
}

// 判断用户是否存在（支持邮箱或钱包地址）
func (s *UserService) UserExists(email, walletAddr string) (bool, error) {
	return s.Repo.UserExists(email, walletAddr)
}
