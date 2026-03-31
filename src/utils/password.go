package utils

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

const maxBcryptPasswordBytes = 72

// HashPassword 生成密码哈希值。
func HashPassword(password string) (string, error) {
	passwordBytes := []byte(password)
	if len(passwordBytes) > maxBcryptPasswordBytes {
		return "", errors.New("password exceeds maximum length of 72 bytes for bcrypt")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
	return string(hashedPassword), err
}

// CheckPassword 校验明文密码与哈希是否匹配。
func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
