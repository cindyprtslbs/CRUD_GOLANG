package repository

import (
	"context"
	"errors"
	"fmt"

	models "crud-app/app/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Interface AuthRepository mendefinisikan kontrak fungsi yang bisa digunakan oleh service
type AuthRepository interface {
	GetByUsername(ctx context.Context, username string) (*models.User, string, error)
}

// Struktur utama repository
type authRepository struct {
	collection *mongo.Collection
}

// Fungsi konstruktor untuk inisialisasi repository
func NewAuthRepository(database *mongo.Database) AuthRepository {
	return &authRepository{
		collection: database.Collection("users"), // ganti sesuai nama collection kamu
	}
}

// GetByUsername mencari user berdasarkan username atau email
func (r *authRepository) GetByUsername(ctx context.Context, username string) (*models.User, string, error) {
	var user models.User

	// Filter: mencari username atau email yang aktif
	filter := bson.M{
		"$and": []bson.M{
			{
				"$or": []bson.M{
					{"username": username},
					{"email": username},
				},
			},
			{"is_active": true},
		},
	}

	// Ambil data dari MongoDB
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, "", fmt.Errorf("user tidak ditemukan")
		}
		return nil, "", err
	}

	// Ambil password hash dari field PasswordHash (pastikan field ini ada di model)
	passwordHash := user.PasswordHash
	if passwordHash == "" {
		return nil, "", fmt.Errorf("password hash tidak ditemukan di data user")
	}

	return &user, passwordHash, nil
}
