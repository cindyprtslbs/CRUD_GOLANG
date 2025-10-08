package service

import (
	models "crud-app/app/model"
	"crud-app/app/repository"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	_ "github.com/lib/pq"

	"github.com/gofiber/fiber/v2"
)

type AlumniService struct {
	repo repository.AlumniRepository
}

func NewAlumniService(r repository.AlumniRepository) *AlumniService {
	return &AlumniService{repo: r}
}

func (s *AlumniService) GetAll(c *fiber.Ctx) error {
	username, _ := c.Locals("username").(string)
	log.Printf("User %s mengakses GET /api/alumni", username)

	list, err := s.repo.GetAll()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Gagal mengambil data alumni",
		})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    list,
	})
}

func (s *AlumniService) GetByID(c *fiber.Ctx) error {
	username, _ := c.Locals("username").(string)

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "ID tidak valid",
		})
	}

	log.Printf("User %s mengakses GET /api/alumni/%d", username, id)

	alumni, err := s.repo.GetByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Alumni tidak ditemukan",
		})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    alumni,
		"message": "Data alumni berhasil diambil",
	})
}

func (s *AlumniService) Create(c *fiber.Ctx) error {
	username, _ := c.Locals("username").(string)
	log.Printf("Admin %s menambah alumni baru", username)

	var req models.CreateAlumniRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Request body tidak valid",
		})
	}

	// Validasi input
	if req.NIM == "" || req.Nama == "" || req.Jurusan == "" || req.Email == "" || req.TahunLulus == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "Semua field harus diisi",
		})
	}

	alumni, err := s.repo.Create(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Gagal menambah alumni baru",
		})
	}
	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"data":    alumni,
		"message": "Alumni berhasil ditambahkan",
	})
}

func (s *AlumniService) Update(c *fiber.Ctx) error {
	username, _ := c.Locals("username").(string)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "ID tidak valid",
		})
	}

	log.Printf("Admin %s mengupdate alumni ID %d", username, id)

	var req models.UpdateAlumniRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Body request tidak valid",
		})
	}

	// Validasi input
	if req.NIM == "" || req.Nama == "" || req.Jurusan == "" || req.Email == "" || req.TahunLulus == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "Semua field harus diisi",
		})
	}

	alumni, err := s.repo.Update(id, req)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{
				"error": "Alumni tidak ditemukan",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Gagal update alumni",
		})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    alumni,
		"message": "Alumni berhasil diupdate",
	})
}

// =================== SOFT DELETE ===================
func (s *AlumniService) SoftDelete(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	role := c.Locals("role").(string)
	userLoginID := c.Locals("user_id").(int)

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID tidak valid"})
	}

	log.Printf("%s mencoba menghapus alumni ID %d", username, id)

	// Cek data alumni
	alumni, err := s.repo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Alumni tidak ditemukan"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menghapus alumni"})
	}

	// Pastikan alumni punya user_id
	if alumni.UserID == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Alumni ini tidak memiliki user terkait, tidak dapat dihapus"})
	}

	// Validasi: hanya admin atau alumni pemilik sendiri yang boleh hapus
	if role != "admin" && userLoginID != *alumni.UserID {
		return c.Status(403).JSON(fiber.Map{"error": "Anda tidak memiliki izin untuk menghapus data ini"})
	}

	// Eksekusi soft delete di repository
	err = s.repo.SoftDelete(id, *alumni.UserID, role)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Gagal menghapus alumni: %v", err)})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Alumni berhasil dihapus (soft delete)",
	})
}

// =================== RESTORE ===================
func (s *AlumniService) Restore(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID tidak valid"})
	}
	log.Printf("User %s merestore alumni ID %d", username, id)

	// Ambil data alumni
	alumni, err := s.repo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Alumni tidak ditemukan"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Gagal merestore alumni"})
	}

	if alumni.UserID == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Data alumni ini tidak memiliki user terkait"})
	}

	err = s.repo.Restore(id, *alumni.UserID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Gagal merestore alumni: %v", err)})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Alumni berhasil direstore dan user diaktifkan kembali",
	})
}

// func (s *AlumniService) Delete(c *fiber.Ctx) error {
// 	username, _ := c.Locals("username").(string)
// 	id, err := strconv.Atoi(c.Params("id"))
// 	if err != nil {
// 		return c.Status(400).JSON(fiber.Map{
// 			"error": "ID tidak valid",
// 		})
// 	}

// 	log.Printf("Admin %s menghapus alumni ID %d", username, id)

// 	err = s.repo.Delete(id)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return c.Status(404).JSON(fiber.Map{
// 				"error": "Alumni tidak ditemukan",
// 			})
// 		}
// 		return c.Status(500).JSON(fiber.Map{
// 			"error": "Gagal menghapus alumni",
// 		})
// 	}
// 	return c.JSON(fiber.Map{
// 		"success": true,
// 		"message": "Alumni berhasil dihapus",
// 	})
// }

func (s *AlumniService) GetWithoutPekerjaan(c *fiber.Ctx) error {
	username, _ := c.Locals("username").(string)
	log.Printf("User %s mengakses GET /api/alumni/without-pekerjaan", username)

	count, err := s.repo.CountWithoutPekerjaan()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Gagal menghitung alumni tanpa pekerjaan",
		})
	}

	list, err := s.repo.GetWithoutPekerjaan()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Gagal mengambil data alumni tanpa pekerjaan",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"status":  "Data Alumni tanpa pekerjaan",
		"count":   count,
		"data":    list,
	})
}

func (s *AlumniService) GetAlumniService(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	sortBy := c.Query("sortBy", "id")
	order := c.Query("order", "asc")
	search := c.Query("search", "")

	offset := (page - 1) * limit

	sortByWhitelist := map[string]bool{
		"id": true, "nim": true, "nama": true, "jurusan": true, "angkatan": true,
		"tahun_lulus": true, "email": true, "created_at": true,
	}
	if !sortByWhitelist[sortBy] {
		sortBy = "id"
	}
	if strings.ToLower(order) != "desc" {
		order = "asc"
	}

	alumni, err := s.repo.GetAlumniRepo(search, sortBy, order, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch alumni"})
	}

	total, err := s.repo.CountAlumniRepo(search)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to count alumni"})
	}

	response := models.AlumniResponse{
		Data: alumni,
		Meta: models.MetaInfo{
			Page:   page,
			Limit:  limit,
			Total:  total,
			Pages:  (total + limit - 1) / limit,
			SortBy: sortBy,
			Order:  order,
			Search: search,
		},
	}

	return c.JSON(response)
}
