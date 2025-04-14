package user

import (
	"AvitoPVZ/internal/config"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

func GenerateToken(userID uuid.UUID, role Role) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		Role:   role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, _ := token.SignedString(os.Getenv(config.JwtSecret))
	fmt.Println("Token generated: ", signedToken)
	return token.SignedString(os.Getenv(config.JwtSecret))
}
