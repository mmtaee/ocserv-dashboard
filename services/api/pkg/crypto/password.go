package crypto

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/mmtaee/ocserv-users-management/common/pkg/config"
	"math/rand"
)

type CustomPassword struct {
	Salt string
	Hash string
}

type CustomPasswordInterface interface {
	CreatePassword(passwd string, saltLength ...int) CustomPassword
	CheckPassword(passwd, hashedPassword, salt string) bool
}

func NewCustomPassword() *CustomPassword {
	return &CustomPassword{}
}

func (c *CustomPassword) CreatePassword(passwd string, saltLength ...int) CustomPassword {
	length := 6
	if len(saltLength) > 0 {
		length = saltLength[0]
	}
	s := salt(length)
	hash := create(passwd, s)

	return CustomPassword{
		Salt: s,
		Hash: hash,
	}
}

func (c *CustomPassword) CheckPassword(passwd, hashedPassword, salt string) bool {
	hash := create(passwd, salt)
	if hashedPassword == hash {
		return true
	}
	return false
}

func salt(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func create(passwd, salt string) string {
	secretKey := config.Get().SecretKey
	passwordHash := fmt.Sprintf("%s%s%s", salt, passwd, secretKey)
	hash := md5.New()
	hash.Write([]byte(passwordHash))
	return hex.EncodeToString(hash.Sum(nil))
}
