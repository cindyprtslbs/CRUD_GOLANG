package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Struktur utama user
type User struct {
	ID           primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	Username     string              `bson:"username" json:"username"`
	Email        string              `bson:"email" json:"email"`
	PasswordHash string              `bson:"password_hash" json:"password_hash"`
	Role         string              `bson:"role" json:"role"`
	IsDeleted    bool                `bson:"is_deleted" json:"is_deleted"`
	IsActive     bool                `bson:"is_active" json:"is_active"`
	AlumniID     *primitive.ObjectID `bson:"alumni_id,omitempty" json:"alumni_id,omitempty"`
	CreatedAt    time.Time           `bson:"created_at" json:"created_at"`
}

// Request login dari client
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Response login ke client
type LoginResponse struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}

// Struktur data di dalam JWT
type JWTClaims struct {
	UserID   primitive.ObjectID `json:"user_id"`
	Username string             `json:"username"`
	Role     string             `json:"role"`
	AlumniID string             `json:"alumni_id,omitempty"`
	jwt.RegisteredClaims
}
