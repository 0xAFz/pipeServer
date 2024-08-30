package services

import (
	"pipe/internal/entity"
	"pipe/internal/repositories"
)

type AccountService struct {
	repo repositories.Account
}

func NewAccountService(repo repositories.Account) *AccountService {
	return &AccountService{repo: repo}
}

func (s *AccountService) GetUserByID(ID int64) (entity.User, error) {
	return s.repo.ByID(ID)
}

func (s *AccountService) GetUserByPrivateID(ID string) (entity.User, error) {
	return s.repo.ByPrivateID(ID)
}

func (s *AccountService) CreateUser(user entity.User) error {
	return s.repo.Save(user)
}

func (s *AccountService) DeleteUser(user entity.User) error {
	return s.repo.Delete(user)
}

func (s *AccountService) SetPubKey(user entity.User) error {
	return s.repo.SetPubKey(user)
}
