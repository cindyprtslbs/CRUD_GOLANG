package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	models "crud-app/app/model"
	"crud-app/app/repository"

	"github.com/gofiber/fiber/v2"
)

type AlumniService struct {
	repo repository.AlumniRepository
}

func NewAlumniService(r repository.AlumniRepository) *AlumniService {
	return &AlumniService{repo: r}
}

// GetAll godoc
// @Summary Mendapatkan semua data alumni
// @Description Mengambil seluruh data alumni dari database tanpa pagination
// @Tags Alumni
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "success response dengan array data alumni"
// @Failure 500 {object} map[string]interface{} "error response"
// @Security Bearer
// @Router /unair/alumni/all [get]
func (s *AlumniService) GetAll(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	username, _ := c.Locals("username").(string)
	log.Printf("User %s mengakses GET /api/alumni", username)

	list, err := s.repo.GetAll(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data alumni"})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    list,
	})
}

// GetByID godoc
// @Summary Mendapatkan data alumni berdasarkan ID
// @Description Mengambil data alumni tertentu berdasarkan ID
// @Tags Alumni
// @Accept json
// @Produce json
// @Param id path string true "ID Alumni (MongoDB ObjectID)"
// @Success 200 {object} map[string]interface{} "success response dengan data alumni"
// @Failure 404 {object} map[string]interface{} "alumni tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "error response"
// @Security Bearer
// @Router /unair/alumni/{id} [get]
func (s *AlumniService) GetByID(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	username, _ := c.Locals("username").(string)
	id := c.Params("id")
	log.Printf("User %s mengakses GET /api/alumni/%s", username, id)

	alumni, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data alumni"})
	}
	if alumni == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Alumni tidak ditemukan"})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    alumni,
	})
}

// Create godoc
// @Summary Menambah data alumni baru
// @Description Menambahkan data alumni ke dalam database (hanya admin)
// @Tags Alumni
// @Accept json
// @Produce json
// @Param body body models.CreateAlumniRequest true "Data Alumni Baru (NIM, Nama, Jurusan, Email, TahunLulus)"
// @Success 201 {object} map[string]interface{} "success response dengan data alumni baru"
// @Failure 400 {object} map[string]interface{} "request body tidak valid atau field kosong"
// @Failure 500 {object} map[string]interface{} "error response"
// @Security Bearer
// @Router /unair/alumni [post]
func (s *AlumniService) Create(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	username, _ := c.Locals("username").(string)
	log.Printf("Admin %s menambah alumni baru", username)

	var req models.CreateAlumniRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Request body tidak valid"})
	}

	if req.NIM == "" || req.Nama == "" || req.Jurusan == "" || req.Email == "" || req.TahunLulus == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Semua field harus diisi"})
	}

	alumni, err := s.repo.Create(ctx, &req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menambah alumni baru"})
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"data":    alumni,
		"message": "Alumni berhasil ditambahkan",
	})
}

// Update godoc
// @Summary Mengupdate data alumni
// @Description Mengubah data alumni berdasarkan ID (hanya admin)
// @Tags Alumni
// @Accept json
// @Produce json
// @Param id path string true "ID Alumni (MongoDB ObjectID)"
// @Param body body models.UpdateAlumniRequest true "Data Alumni yang Diupdate (minimal NIM, Nama, Jurusan, Email, TahunLulus)"
// @Success 200 {object} map[string]interface{} "success response dengan data alumni yang diupdate"
// @Failure 400 {object} map[string]interface{} "request body tidak valid atau field kosong"
// @Failure 404 {object} map[string]interface{} "alumni tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "error response"
// @Security Bearer
// @Router /unair/alumni/{id} [put]
func (s *AlumniService) Update(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	username, _ := c.Locals("username").(string)
	id := c.Params("id")
	log.Printf("Admin %s mengupdate alumni ID %s", username, id)

	var req models.UpdateAlumniRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Body request tidak valid"})
	}

	if req.NIM == "" || req.Nama == "" || req.Jurusan == "" || req.Email == "" || req.TahunLulus == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Semua field harus diisi"})
	}

	alumni, err := s.repo.Update(ctx, id, &req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal update alumni"})
	}
	if alumni == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Alumni tidak ditemukan"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    alumni,
		"message": "Alumni berhasil diupdate",
	})
}

// SoftDelete godoc
// @Summary Menghapus alumni (soft delete)
// @Description Mengubah status alumni menjadi nonaktif tanpa menghapus data dari database
// @Tags Alumni
// @Accept json
// @Produce json
// @Param id path string true "ID Alumni (MongoDB ObjectID)"
// @Success 200 {object} map[string]interface{} "success response dengan message"
// @Failure 404 {object} map[string]interface{} "alumni tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "error response"
// @Security Bearer
// @Router /unair/alumni/{id} [delete]
func (s *AlumniService) SoftDelete(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	username := c.Locals("username").(string)
	id := c.Params("id")

	log.Printf("%s mencoba menghapus alumni ID %s", username, id)

	alumni, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mencari alumni"})
	}
	if alumni == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Alumni tidak ditemukan"})
	}

	err = s.repo.SoftDelete(ctx, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Gagal menghapus alumni: %v", err)})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Alumni berhasil dihapus (soft delete)",
	})
}

// Restore godoc
// @Summary Merestore alumni yang dihapus
// @Description Mengembalikan alumni yang sebelumnya dihapus (soft delete) menjadi aktif kembali
// @Tags Alumni
// @Accept json
// @Produce json
// @Param id path string true "ID Alumni (MongoDB ObjectID)"
// @Success 200 {object} map[string]interface{} "success response dengan message"
// @Failure 500 {object} map[string]interface{} "error response"
// @Security Bearer
// @Router /unair/alumni/{id} [patch]
func (s *AlumniService) Restore(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	username := c.Locals("username").(string)
	id := c.Params("id")

	log.Printf("User %s merestore alumni ID %s", username, id)

	err := s.repo.Restore(ctx, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Gagal merestore alumni: %v", err)})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Alumni berhasil direstore",
	})
}

// GetWithoutPekerjaan godoc
// @Summary Mendapatkan alumni tanpa pekerjaan
// @Description Mengambil daftar alumni yang belum memiliki pekerjaan dengan jumlah totalnya
// @Tags Alumni
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "success response dengan array data alumni tanpa pekerjaan"
// @Failure 500 {object} map[string]interface{} "error response"
// @Security Bearer
// @Router /unair/alumni/without-pekerjaan [get]
func (s *AlumniService) GetWithoutPekerjaan(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	username, _ := c.Locals("username").(string)
	log.Printf("User %s mengakses GET /api/alumni/without-pekerjaan", username)

	count, err := s.repo.CountWithoutPekerjaan(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menghitung alumni tanpa pekerjaan"})
	}

	list, err := s.repo.GetWithoutPekerjaan(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data alumni tanpa pekerjaan"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"count":   count,
		"data":    list,
	})
}

// GetAlumniService godoc
// @Summary Mendapatkan daftar alumni dengan pencarian, sorting, dan pagination
// @Description Mengambil daftar alumni berdasarkan parameter pencarian, pengurutan, dan batas halaman dengan informasi meta
// @Tags Alumni
// @Accept json
// @Produce json
// @Param page query int false "Nomor halaman (default: 1)"
// @Param limit query int false "Jumlah data per halaman (default: 10)"
// @Param sortBy query string false "Kolom untuk sorting: nim, nama, jurusan, angkatan, tahun_lulus, email, created_at (default: nama)"
// @Param order query string false "Urutan sorting: asc atau desc (default: asc)"
// @Param search query string false "Kata kunci pencarian di semua field"
// @Success 200 {object} models.AlumniResponse "success response dengan data alumni dan meta informasi"
// @Failure 500 {object} map[string]interface{} "error response"
// @Security Bearer
// @Router /unair/alumni [get]
func (s *AlumniService) GetAlumniService(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	sortBy := c.Query("sortBy", "nama")
	order := c.Query("order", "asc")
	search := c.Query("search", "")

	offset := int64((page - 1) * limit)

	sortByWhitelist := map[string]bool{
		"nim": true, "nama": true, "jurusan": true, "angkatan": true,
		"tahun_lulus": true, "email": true, "created_at": true,
	}
	if !sortByWhitelist[sortBy] {
		sortBy = "nama"
	}
	if strings.ToLower(order) != "desc" {
		order = "asc"
	}

	alumni, err := s.repo.GetAlumniRepo(ctx, search, sortBy, order, int64(limit), offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data alumni"})
	}

	total, err := s.repo.CountAlumniRepo(ctx, search)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menghitung total alumni"})
	}

	response := models.AlumniResponse{
		Data: alumni,
		Meta: models.MetaInfo{
			Page:   page,
			Limit:  limit,
			Total:  int(total),
			Pages:  int((total + int64(limit) - 1) / int64(limit)),
			SortBy: sortBy,
			Order:  order,
			Search: search,
		},
	}

	return c.JSON(response)
}