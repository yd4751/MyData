package repository

import (
	"mediastation/internal/model"
	"mediastation/pkg/database"
)

type UserRepository interface {
	GetByID(id uint) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
	Create(user *model.User) error
	Update(user *model.User) error
	Delete(id uint) error
}

type userRepository struct{}

func NewUserRepository() UserRepository {
	return &userRepository{}
}

func (r *userRepository) GetByID(id uint) (*model.User, error) {
	var user model.User
	err := database.DB.First(&user, id).Error
	return &user, err
}

func (r *userRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := database.DB.Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *userRepository) Create(user *model.User) error {
	return database.DB.Create(user).Error
}

func (r *userRepository) Update(user *model.User) error {
	return database.DB.Save(user).Error
}

func (r *userRepository) Delete(id uint) error {
	return database.DB.Delete(&model.User{}, id).Error
}
