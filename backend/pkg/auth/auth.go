package auth

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes the password using bcrypt with a generated salt
func HashPassword(password []byte) ([]byte, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return hashedPassword, nil
}

// CheckPasswordHash compares the hashed password with the plain password
func CheckPasswordHash(password, hashedPassword []byte) bool {
	err := bcrypt.CompareHashAndPassword(hashedPassword, password)
	return err == nil
}

// GenerateRandomString generates a random string of the given length
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.IntN(len(charset))] //nolint:gosec // it's okay to use math/rand here
	}
	return string(b)
}

func CreateJWT(userID, issuer, secret string) (string, error) {
	signingKey := []byte(secret)

	claims := &jwt.RegisteredClaims{
		Issuer:    issuer,
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}

	return ss, nil
}

func CheckJWT(bearerToken, issuer, jwtSecret string) (string, error) {
	// Parse the token
	token, err := jwt.Parse(bearerToken, func(token *jwt.Token) (interface{}, error) {
		// Make sure that the token method conforms to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		signingKey := []byte(jwtSecret)
		return signingKey, nil
	})
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, err := claims.GetSubject()
		if err != nil {
			return "", fmt.Errorf("invalid subject: %w", err)
		}
		actualIssuer, err := claims.GetIssuer()
		if err != nil || actualIssuer != issuer {
			return "", fmt.Errorf("invalid issuer: %w", err)
		}

		return userID, nil
	}

	return "", errors.New("invalid token")
}
