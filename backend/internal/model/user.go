package repository

import (
    "gorm.io/gorm"
)

type User struct {
    ID       uint   `gorm:"primaryKey"`
    Email    string `gorm:"uniqueIndex;not null"`
    Password string `gorm:"not null"`
}

type UserRepository interface {
    CreateUser(user *User) error
    GetUserByEmail(email string) (*User, error)
    GetUserByID(id uint) (*User, error)
}

type userRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
    db.AutoMigrate(&User{})
    return &userRepository{db: db}
}

func (r *userRepository) CreateUser(user *User) error {
    return r.db.Create(user).Error
}

func (r *userRepository) GetUserByEmail(email string) (*User, error) {
    var user User
    if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) GetUserByID(id uint) (*User, error) {
    var user User
    if err := r.db.First(&user, id).Error; err != nil {
        return nil, err
    }
    return &user, nil
}
