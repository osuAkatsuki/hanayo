package auth

import (
	"github.com/osuAkatsuki/hanayo/app/usecases/auth/cryptography"
	tp "github.com/osuAkatsuki/hanayo/internal/top-passwords"
	"golang.org/x/crypto/bcrypt"
)

func CompareHashPasswords(bcryptHash string, password string) error {
	bcryptHashBytes := []byte(bcryptHash)
	passwordHashBytes := []byte(cryptography.MakeMD5(password))
	err := bcrypt.CompareHashAndPassword(bcryptHashBytes, passwordHashBytes)
	return err
}

func ValidatePassword(p string) string {
	if len(p) < 8 {
		return "Your password is too short! It must be at least 8 characters long."
	}

	for _, k := range tp.TopPasswords {
		if k == p {
			return "Your password is one of the most common passwords on the entire internet. No way we're letting you use that!"
		}
	}

	return ""
}

func GeneratePassword(p string) (string, error) {
	s, err := cryptography.MakeBcrypt(cryptography.MakeMD5(p))
	return string(s), err
}
