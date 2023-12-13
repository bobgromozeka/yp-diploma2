package jwt

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

var SecretKey = []byte("6ecafb3785ddd92172f71ec4821211e124e0527a268cb6445c8fdaa02c6a2628f25bfb17ef95f1a0873ab87f6f559958deccb7c7902514f0164efd99950a670c")

var ErrWrongJWT = errors.New("invalid JWT token")

const UserIDKey = "userID"

// NewJWTWithUserID creates JWT for specified user ID
func NewJWTWithUserID(userID int) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256, jwt.MapClaims{
			UserIDKey: userID,
		},
	)

	strToken, signErr := token.SignedString(SecretKey)

	return strToken, signErr
}

// GetClaimsFromSign returns payload from specified jwt
func GetClaimsFromSign(signString string) (jwt.MapClaims, error) {
	token, tokenErr := jwt.Parse(signString, func(token *jwt.Token) (interface{}, error) {
		return SecretKey, nil
	})
	if tokenErr != nil {
		return nil, tokenErr
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, ErrWrongJWT
	}
}
