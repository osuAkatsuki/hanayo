package cryptography

import (
	"crypto/md5"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func MakeMD5(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

func MakeBcrypt(p string) (string, error) {
	s, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(s), err
}
