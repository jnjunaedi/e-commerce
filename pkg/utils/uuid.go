package utils

import (
	"e-commerce/pkg/utils/apperror"

	"github.com/google/uuid"
)

type IDGenerator interface {
	GenerateUUID() (string, error)
}

type idGeneratorImpl struct{}

func NewIDGenerator() IDGenerator {
	return &idGeneratorImpl{}
}

func (i *idGeneratorImpl) GenerateUUID() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", apperror.New(apperror.ErrInternalServer, "error internal server")
	}

	return id.String(), nil
}
