package service

import (
	"context"
	"errors"
	"time"

	models "crud-app/app/model"
	"crud-app/app/repository"
	"crud-app/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthService struct {
	repo repository.AuthRepository
}

func NewAuthService(r repository.AuthRepository) *AuthService {
	return &AuthService{repo: r}
}

// @Summary Login user
// @Description Login dengan username dan password
// @Accept json
// @Tags Auth
// @Produce json
// @Param loginRequest body models.LoginRequest true "Login Request"
// @Success 200 {object} map[string]interface{} "Berhasil mendapatkan semua user"
// @Failure 403 {object} map[string]interface{} "Akses ditolak, bukan admin"
// @Failure 500 {object} map[string]interface{} "Kesalahan server"
// @Security Bearer
// @Router /api/login [post]
func (s *AuthService) Login(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Request body tidak valid",
		})
	}

	// Validasi input
	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username dan password harus diisi",
		})
	}

	// Ambil user dari MongoDB
	user, passwordHash, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) || user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Username atau password salah",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal mengambil data user dari database",
		})
	}

	// Cek password hash
	if !utils.CheckPasswordHash(req.Password, passwordHash) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Username atau password Hash salah",
		})
	}

	// Generate JWT token
	token, err := utils.GenerateToken(*user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal membuat token autentikasi",
		})
	}

	response := models.LoginResponse{
		User:  *user,
		Token: token,
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Login berhasil",
		"data":    response,
	})
}

// @Summary Get user profile
// @Description Get profile dari user yang sedang login
// @Produce json
// @Success 200 {object} map[string]interface{} "Berhasil login dan mengembalikan token"
// @Failure 401 {object} map[string]interface{} "Username/password salah"
// @Security Bearer
// @Router /api/profile [get]
func (s *AuthService) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	username := c.Locals("username")
	role := c.Locals("role")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Profile berhasil diambil",
		"data": fiber.Map{
			"user_id":  userID,
			"username": username,
			"role":     role,
		},
	})
}
