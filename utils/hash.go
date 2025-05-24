package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
)

var privateKey = []byte(os.Getenv("SECRET_KEY"))

func HashAString(stringToHash string) (string, error) {
	h := hmac.New(sha256.New, privateKey)

	_, err := h.Write([]byte(stringToHash))
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func HashAndCompareTwoStrings(stringToHash, stringToCompare string) (bool, error) {
	hashedString, err := HashAString(stringToHash)
	if err != nil {
		return false, err
	}
	return hmac.Equal([]byte(hashedString), []byte(stringToCompare)), nil
}

func CreateJWTToken(claims jwt.MapClaims) (string, error) {
	var expStr = os.Getenv("EXPIRATION_TIME")
	var expiration int
	if expStr != "" {
		expiration, _ = strconv.Atoi(expStr)
	}

	claims["exp"] = time.Now().Add(time.Duration(expiration) * time.Second).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(privateKey)
}

func DecodeJWTToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return privateKey, nil
	})
}
