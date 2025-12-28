package mocks

import "github.com/stretchr/testify/mock"

type Hasher struct {
	mock.Mock
}

func NewHasher() *Hasher {
	return &Hasher{}
}

func (m *Hasher) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *Hasher) CompareHashAndPassword(hashed, plain string) error {
	args := m.Called(hashed, plain)
	return args.Error(0)
}
