package services

import (
	"crypto/rand"

	"github.com/unshade/citadelle/internal/config"
	"golang.org/x/crypto/argon2"
)

func GenerateMasterKey() ([]byte, error) {
	key := make([]byte, config.KeySize)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func DeriveKEK(password string, salt []byte) ([]byte, error) {
	key := argon2.IDKey(
		[]byte(password),
		salt,
		1,
		64*1024,
		4,
		config.KeySize,
	)
	return key, nil
}
