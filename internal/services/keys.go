package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
)

func EncryptKey(key, encryptionKey []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(encryptionKey)
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

	ciphertext := gcm.Seal(nonce, nonce, key, nil)

	return ciphertext, nonce, nil
}

func DecryptKey(encryptedData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
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
