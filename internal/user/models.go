package user

import (
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type Role string

const (
	RoleModerator Role = "moderator"
	RoleEmployee  Role = "employee"
	RoleUnknown   Role = "unknown"
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   Role      `json:"role"`
	jwt.StandardClaims
}

type User struct {
	ID        uuid.UUID
	Email     string
	HashedPwd string
	Role      Role
}

type RegisterUserRequest struct {
	Email    string
	Password string
	Role     Role
}

type LoginRequest struct {
	Email    string
	Password string
}

type InsertUserQuery struct {
	Email     string
	HashedPwd string
	Role      Role
}

type GetUserQuery struct {
	Email string
}
