package jwt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const expectedJWT = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySUQiOjF9.fXMEMQ5FGglGA64ijzTPDf-94QSY_Bv1KLcbkHmPO1s"

func TestNewJWTWithUserID(t *testing.T) {
	userID := 1
	token, err := NewJWTWithUserID(userID)

	assert.Nil(t, err)
	assert.Equal(t, expectedJWT, token)
}

func TestGetClaimsFromSign(t *testing.T) {
	claims, err := GetClaimsFromSign(expectedJWT)

	assert.Nil(t, err)
	assert.Equal(t, float64(1), claims["userID"])
}
