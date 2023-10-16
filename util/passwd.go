package util

import "golang.org/x/crypto/bcrypt"

func GetHashedPasswd(passwd string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(passwd), 10)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

func CheckPasswd(hashedPasswd, passwd string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPasswd), []byte(passwd)) == nil
}
