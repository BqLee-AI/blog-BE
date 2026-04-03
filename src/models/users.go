package models

import (
	"blog-BE/src/dao"
)

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `gorm:"not null" json:"username"`
	Password string `gorm:"not null" json:"-"`
	Email    string `gorm:"unique;not null" json:"email"`
	RoleID   uint   `gorm:"not null" json:"role_id"` // 0: 普通用户, 1: 管理员, 2: 超级管理员
}

func (User) TableName() string {
	return "users" // 返回你想要的精确表名
}

/*
 * User 的增删查改
 */
func CreateUser(user *User) error {
	if err := dao.DB.Create(user).Error; err != nil {
		return err
	}
	return nil
}

func FindUserByID(id uint) (*User, error) {
	var user User
	if err := dao.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func FindUserByUsername(username string) (*User, error) {
	var user User
	if err := dao.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func FindUserByEmail(email string) (*User, error) {
	var user User
	if err := dao.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func UpdateUser(user *User) error {
	if err := dao.DB.Save(user).Error; err != nil {
		return err
	}
	return nil
}

func DeleteUser(id uint) error {
	if err := dao.DB.Delete(&User{}, id).Error; err != nil {
		return err
	}
	return nil
}
