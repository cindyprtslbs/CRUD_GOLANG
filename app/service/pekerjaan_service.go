package service

import (
	models "crud-app/app/model"
	"crud-app/app/repository"
	"database/sql"
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type PekerjaanService struct {
	repo repository.PekerjaanRepository
}

func NewPekerjaanService(r repository.PekerjaanRepository) *PekerjaanService {
	return &PekerjaanService{repo: r}
}

// =================== GET ===================
func (s *PekerjaanService) GetAll(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	log.Printf("User %s mengakses GET /api/pekerjaan", username)

	list, err := s.repo.GetAll()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data pekerjaan"})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    list,
	})
}

func (s *PekerjaanService) GetByID(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID tidak valid"})
	}

	log.Printf("User %s mengakses GET /api/pekerjaan/%d", username, id)

	p, err := s.repo.GetByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Pekerjaan tidak ditemukan"})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    p,
	})
}

func (s *PekerjaanService) GetByAlumniID(c *fiber.Ctx) error {
	username := c.Locals("username").(string)

	alumniID, err := strconv.Atoi(c.Params("alumni_id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID Alumni tidak valid"})
	}

	log.Printf("User %s mengakses GET /api/pekerjaan/alumni/%d", username, alumniID)

	list, err := s.repo.GetByAlumniID(alumniID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data pekerjaan alumni"})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    list,
	})
}

// =================== CREATE ===================
func (s *PekerjaanService) Create(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	log.Printf("Admin %s menambah pekerjaan_alumni baru", username)

	var req models.CreatePekerjaanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Body request tidak valid"})
	}

	// Validasi input
	if req.AlumniID == 0 || req.NamaPerusahaan == "" || req.PosisiJabatan == "" ||
		req.BidangIndustri == "" || req.LokasiKerja == "" || req.GajiRange == "" ||
		req.TanggalMulaiKerja == "" || req.StatusPekerjaan == "" || req.DeskripsiPekerjaan == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Semua field wajib diisi"})
	}

	newPekerjaan, err := s.repo.Create(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menambah pekerjaan"})
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"data":    newPekerjaan,
	})
}

// =================== UPDATE ===================
func (s *PekerjaanService) Update(c *fiber.Ctx) error {
	username := c.Locals("username").(string)

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID tidak valid"})
	}

	log.Printf("Admin %s mengupdate pekerjaan_alumni ID %d", username, id)

	var req models.UpdatePekerjaanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Body request tidak valid"})
	}

	if req.NamaPerusahaan == "" || req.PosisiJabatan == "" ||
		req.BidangIndustri == "" || req.LokasiKerja == "" || req.GajiRange == "" ||
		req.TanggalMulaiKerja == "" || req.StatusPekerjaan == "" || req.DeskripsiPekerjaan == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Semua field wajib diisi"})
	}

	updated, err := s.repo.Update(id, req)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Pekerjaan tidak ditemukan"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Gagal update pekerjaan"})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    updated,
	})
}

// =================== DELETE ===================

func (s *PekerjaanService) SoftDelete(c *fiber.Ctx) error {
	role := c.Locals("role")
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID tidak valid"})
	}

	roleStr, ok := role.(string)
	if !ok {
		return c.Status(403).JSON(fiber.Map{"error": "Role tidak dikenali"})
	}

	// admin bisa hapus semua data
	if roleStr == "admin" {
		if err := s.repo.SoftDeleteByID(id); err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).JSON(fiber.Map{"error": "Pekerjaan tidak ditemukan"})
			}
			return c.Status(500).JSON(fiber.Map{"error": "Gagal soft delete pekerjaan"})
		}
		return c.JSON(fiber.Map{"success": true, "message": "Pekerjaan berhasil dihapus oleh admin"})
	}

	// user biasa (alumni)
	userID := c.Locals("user_id")
	var uid int
	switch v := userID.(type) {
	case int:
		uid = v
	case float64:
		uid = int(v)
	default:
		return c.Status(403).JSON(fiber.Map{"error": "User tidak valid"})
	}

	// Ambil user dari DB
	user, err := s.repo.GetUserByID(uid)
	if err != nil || user.AlumniID == nil {
		return c.Status(403).JSON(fiber.Map{"error": "User tidak terkait alumni"})
	}

	// Pastikan role benar-benar alumni
	if user.Role != "alumni" {
		return c.Status(403).JSON(fiber.Map{"error": "Hanya alumni yang boleh hapus pekerjaan miliknya"})
	}

	// Ambil data pekerjaan dulu
	pekerjaan, err := s.repo.GetByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Pekerjaan tidak ditemukan"})
	}

	// Validasi kepemilikan
	if pekerjaan.AlumniID != *user.AlumniID {
		return c.Status(403).JSON(fiber.Map{"error": "Kamu tidak boleh menghapus pekerjaan milik orang lain"})
	}

	// Kalau lolos semua → hapus
	if err := s.repo.SoftDeleteByID(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal soft delete pekerjaan"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Pekerjaanmu berhasil dihapus"})

}

// =================== LAPORAN ===================
func (s *PekerjaanService) GetBekerjalebih1Tahun(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	log.Printf("User %s mengakses GET /api/pekerjaan/bekerja-lebih-setahun", username)

	list, err := s.repo.GetBekerjalebih1Tahun()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data"})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"count":   len(list),
		"data":    list,
	})
}

// =================== PAGINATION ===================
func (s *PekerjaanService) GetPekerjaanService(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	sortBy := c.Query("sortBy", "id")
	order := c.Query("order", "asc")
	search := c.Query("search", "")

	offset := (page - 1) * limit

	sortByWhitelist := map[string]bool{
		"id": true, "nama_perusahaan": true, "posisi_jabatan": true,
		"bidang_industri": true, "lokasi_kerja": true, "created_at": true,
	}
	if !sortByWhitelist[sortBy] {
		sortBy = "id"
	}
	if strings.ToLower(order) != "desc" {
		order = "asc"
	}

	pekerjaan, err := s.repo.GetPekerjaanRepo(search, sortBy, order, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch pekerjaan"})
	}

	total, err := s.repo.CountPekerjaanRepo(search)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to count pekerjaan"})
	}

	response := models.PekerjaanResponse{
		Data: pekerjaan,
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

// =================== RESTORE ===================
func (s *PekerjaanService) Restore(c *fiber.Ctx) error {
	role := c.Locals("role")
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID tidak valid"})
	}
	roleStr, ok := role.(string)
	if !ok {
		return c.Status(403).JSON(fiber.Map{"error": "Role tidak dikenali"})
	}

	// admin bisa restore semua data
	if roleStr == "admin" {
		if err := s.repo.RestoreByID(id); err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).JSON(fiber.Map{"error": "Pekerjaan tidak ditemukan"})
			}
			return c.Status(500).JSON(fiber.Map{"error": "Gagal restore pekerjaan"})
		}
		return c.JSON(fiber.Map{"success": true, "message": "Pekerjaan berhasil direstore oleh admin"})
	}

	// user biasa (alumni)
	userID := c.Locals("user_id")
	var uid int
	switch v := userID.(type) {
	case int:
		uid = v
	case float64:
		uid = int(v)
	default:
		return c.Status(403).JSON(fiber.Map{"error": "User tidak valid"})
	}

	// Ambil user dari DB
	user, err := s.repo.GetUserByID(uid)
	if err != nil || user.AlumniID == nil {
		return c.Status(403).JSON(fiber.Map{"error": "User tidak terkait alumni"})
	}
	// Pastikan role benar-benar alumni
	if user.Role != "alumni" {
		return c.Status(403).JSON(fiber.Map{"error": "Hanya alumni yang boleh restore pekerjaan miliknya"})
	}
	// Ambil data pekerjaan dulu
	pekerjaan, err := s.repo.GetByID(id)

	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Pekerjaan tidak ditemukan"})
	}
	// Validasi kepemilikan
	if pekerjaan.AlumniID != *user.AlumniID {
		return c.Status(403).JSON(fiber.Map{"error": "Kamu tidak boleh merestore pekerjaan milik orang lain"})
	}
	// Kalau lolos semua → restore
	if err := s.repo.RestoreByOwner(id, *user.AlumniID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal restore pekerjaan"})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Pekerjaanmu berhasil direstore"})
}

// pekerjaanService.GetTrash admin bisa lihat semua data di trash, alumni hanya bisa lihat miliknya
func (s *PekerjaanService) GetTrash(c *fiber.Ctx) error {
	role := c.Locals("role")
	roleStr, ok := role.(string)
	if !ok {
		return c.Status(403).JSON(fiber.Map{"error": "Role tidak dikenali"})
	}
	// admin bisa lihat semua data di trash
	if roleStr == "admin" {
		list, err := s.repo.GetTrash()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data trash"})
		}
		return c.JSON(fiber.Map{
			"success": true,
			"count":   len(list),
			"data":    list,
		})
	}
	// user biasa (alumni)
	userID := c.Locals("user_id")
	var uid int
	switch v := userID.(type) {
	case int:
		uid = v
	case float64:
		uid = int(v)
	default:
		return c.Status(403).JSON(fiber.Map{"error": "User tidak valid"})
	}
	// Ambil user dari DB
	user, err := s.repo.GetUserByID(uid)
	if err != nil || user.AlumniID == nil {
		return c.Status(403).JSON(fiber.Map{"error": "User tidak terkait alumni"})
	}
	// Pastikan role benar-benar alumni
	if user.Role != "alumni" {
		return c.Status(403).JSON(fiber.Map{"error": "Hanya alumni yang boleh lihat trash pekerjaan miliknya"})
	}
	// Ambil data trash milik alumni
	list, err := s.repo.GetTrashByOwner(*user.AlumniID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data trash milikmu"})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"count":   len(list),
		"data":    list,
	})
}

// HARD DELETE
// ADMIN BISA HARD DELETE SEMUA DATA
// ALUMNI HANYA BISA HARD DELETE MILIKNYA SENDIRI
func (s *PekerjaanService) Delete(c *fiber.Ctx) error {
	role := c.Locals("role")
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID tidak valid"})
	}
	roleStr, ok := role.(string)
	if !ok {
		return c.Status(403).JSON(fiber.Map{"error": "Role tidak dikenali"})
	}
	// admin bisa delete semua data
	if roleStr == "admin" {
		if err := s.repo.Delete(id, nil); err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).JSON(fiber.Map{"error": "Pekerjaan tidak ditemukan"})
			}
			return c.Status(500).JSON(fiber.Map{"error": "Gagal delete pekerjaan"})
		}
		return c.JSON(fiber.Map{"success": true, "message": "Pekerjaan berhasil dihapus oleh admin"})
	}
	// user biasa (alumni)
	userID := c.Locals("user_id")
	var uid int
	switch v := userID.(type) {
	case int:
		uid = v
	case float64:
		uid = int(v)
	default:
		return c.Status(403).JSON(fiber.Map{"error": "User tidak valid"})
	}
	// Ambil user dari DB
	user, err := s.repo.GetUserByID(uid)
	if err != nil || user.AlumniID == nil {
		return c.Status(403).JSON(fiber.Map{"error": "User tidak terkait alumni"})
	}
	// Pastikan role benar-benar alumni
	if user.Role != "alumni" {
		return c.Status(403).JSON(fiber.Map{"error": "Hanya alumni yang boleh hapus pekerjaan miliknya"})
	}
	// Ambil data pekerjaan dulu
	pekerjaan, err := s.repo.GetByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Pekerjaan tidak ditemukan"})
	}
	// Validasi kepemilikan
	if pekerjaan.AlumniID != *user.AlumniID {
		return c.Status(403).JSON(fiber.Map{"error": "Kamu tidak boleh menghapus pekerjaan milik orang lain"})
	}
	// Kalau lolos semua → hapus
	if err := s.repo.Delete(id, user.AlumniID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal delete pekerjaan"})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Pekerjaanmu berhasil dihapus permanen"})
}
