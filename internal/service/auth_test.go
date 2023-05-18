package service

import (
	"github.com/chucky-1/finance/internal/repository/mocks"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAuth_generatePassword(t *testing.T) {
	userRepo := new(mocks.Authorization)
	salt := "iuyuofritu"
	userServ := NewAuth(userRepo, salt)
	inputPassword := "myNewStrongPassword"
	hashPasswordOne := userServ.generatePassword(inputPassword)
	hashPasswordTwo := userServ.generatePassword(inputPassword)
	require.Equal(t, hashPasswordOne, hashPasswordTwo)
	logrus.Infof("hash password: %s", hashPasswordOne)
}
