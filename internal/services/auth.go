package services

import (
	"crypto/rand"
	"errors"

	"github.com/unshade/citadelle/internal/db"
	"github.com/unshade/citadelle/internal/models"
)

func Login(email, password string) (*models.User, []byte, error) {
	user, err := db.GetUserByEmail(email)
	if err != nil {
		return nil, nil, errors.New("identifiants incorrects")
	}

	kek, err := DeriveKEK(password, user.Salt)
	if err != nil {
		return nil, nil, err
	}

	masterKey, err := DecryptKey(user.EncryptedMasterKey, kek)
	if err != nil {
		return nil, nil, errors.New("identifiants incorrects")
	}

	return user, masterKey, nil
}

func CreateUser(email, password string) (*models.User, error) {
	salt := make([]byte, 16)
	rand.Read(salt)

	kek, err := DeriveKEK(password, salt)
	if err != nil {
		return nil, err
	}

	masterKey, err := GenerateMasterKey()
	if err != nil {
		return nil, err
	}

	encryptedMasterKey, _, err := EncryptKey(masterKey, kek)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:              email,
		Salt:               salt,
		EncryptedMasterKey: encryptedMasterKey,
	}

	err = db.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}
