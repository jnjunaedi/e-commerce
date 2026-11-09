package security

import (
	"e-commerce/pkg/utils/apperror"

	"golang.org/x/crypto/bcrypt"
)

type Password interface {
	HashPassword(password string) (string, error)
	ComparePassword(plainPassword string, hashPassword string) bool
}

type PasswordImpl struct{}

func NewPassword() Password {
	return &PasswordImpl{}
}

func (s *PasswordImpl) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", apperror.New(apperror.ErrInternalServer, "error internal server")
	}

	return string(bytes), nil
}

func (s *PasswordImpl) ComparePassword(plainPassword string, hashPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(plainPassword))
	return err == nil
}
