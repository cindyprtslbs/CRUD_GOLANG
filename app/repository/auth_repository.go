package repository

import (
	models "crud-app/app/model"
	"database/sql"
	"fmt"
)

type AuthRepository interface {
	GetByUsername(username string) (*models.User, string, error)
}

type authRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) GetByUsername(username string) (*models.User, string, error) {
	var user models.User
	var passwordHash string

	err := r.db.QueryRow(`
		SELECT id, username, email, password_hash, role, alumni_id, created_at, is_active
		FROM users
		WHERE (username = $1 OR email = $1)
		  AND is_active = TRUE
	`, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&passwordHash,
		&user.Role,
		&user.AlumniID,
		&user.CreatedAt,
		&user.IsActive,
	)

	if err != nil {
		fmt.Println("GetByUsername error:", err)
		return nil, "", err
	}

	return &user, passwordHash, nil
}
