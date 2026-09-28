package auth

import (
	"fmt"
	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string,error) {
	hash,err := argon2id.CreateHash(password, argon2id.DefaultParams)

	if err != nil {
		return "", err
	}

	return hash, nil

}

func CheckPasswordHash(password, hash string) (bool,error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)

	fmt.Println(match)


	if err != nil {
		return false, err
	}


	return match, nil

}
