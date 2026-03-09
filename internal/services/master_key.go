package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"

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

func EncryptMasterKey(masterKey, kek []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, masterKey, nil)

	return ciphertext, nonce, nil
}

func DecryptMasterKey(encryptedData, kek []byte) ([]byte, error) {
	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, errors.New("données chiffrées trop courtes")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]

	masterKey, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("échec déchiffrement (mot de passe incorrect ?)")
	}

	return masterKey, nil
}
