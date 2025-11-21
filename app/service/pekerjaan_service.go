package service

import (
	"context"
	"errors"
	"log"

	models "crud-app/app/model"
	"crud-app/app/repository"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

type PekerjaanService struct {
	repo repository.PekerjaanRepository
}

func NewPekerjaanService(r repository.PekerjaanRepository) *PekerjaanService {
	return &PekerjaanService{repo: r}
}

// @Summary Get all pekerjaan
// @Description Get daftar semua data pekerjaan alumni
// @Tags Pekerjaan_Alumni
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security Bearer
// @Router /unair/pekerjaan-alumni [get]
func (s *PekerjaanService) GetAll(c *fiber.Ctx) error {
	ctx := context.Background()
	username := c.Locals("username").(string)
	log.Printf("User %s mengakses GET /api/pekerjaan", username)

	data, err := s.repo.GetAll(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if len(data) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Data pekerjaan tidak ditemukan"})
	}
	return c.JSON(fiber.Map{"success": true, "data": data})
}

// @Summary Get pekerjaan by ID
// @Description Get detail data pekerjaan berdasarkan ID
// @Tags Pekerjaan_Alumni
// @Produce json
// @Param id path string true "Pekerjaan ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security Bearer
// @Router /unair/pekerjaan-alumni/{id} [get]
func (s *PekerjaanService) GetByID(c *fiber.Ctx) error {
	ctx := context.Background()
	id := c.Params("id")

	data, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return c.Status(404).JSON(fiber.Map{"error": "Pekerjaan tidak ditemukan"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "data": data})
}

// @Summary Get pekerjaan by alumni ID
// @Description Get data pekerjaan berdasarkan alumni ID (admin only)
// @Tags Pekerjaan_Alumni
// @Produce json
// @Param alumni_id path string true "Alumni ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security Bearer
// @Router /unair/pekerjaan-alumni/alumni/{alumni_id} [get]
func (s *PekerjaanService) GetByAlumniID(c *fiber.Ctx) error {
	ctx := context.Background()
	alumniID := c.Params("alumni_id")

	data, err := s.repo.GetByAlumniID(ctx, alumniID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if len(data) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Data pekerjaan alumni tidak ditemukan"})
	}
	return c.JSON(fiber.Map{"success": true, "data": data})
}

// @Summary Create new pekerjaan
// @Description Tambah data pekerjaan alumni baru (admin only)
// @Accept json
// @Tags Pekerjaan_Alumni
// @Produce json
// @Param pekerjaanRequest body models.CreatePekerjaanRequest true "Create Pekerjaan Request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security Bearer
// @Router /unair/pekerjaan-alumni [post]
func (s *PekerjaanService) Create(c *fiber.Ctx) error {
	ctx := context.Background()
	username := c.Locals("username").(string)
	log.Printf("Admin %s menambahkan data pekerjaan_alumni baru", username)

	// Gunakan struct CreatePekerjaanRequest (bukan Pekerjaan langsung)
	var req models.CreatePekerjaanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Body request tidak valid",
			"hint":  "Pastikan format JSON sesuai dan semua field string",
		})
	}

	// Validasi field wajib
	if req.AlumniID == "" || req.NamaPerusahaan == "" || req.PosisiJabatan == "" ||
		req.BidangIndustri == "" || req.LokasiKerja == "" || req.GajiRange == "" ||
		req.TanggalMulaiKerja == "" || req.StatusPekerjaan == "" || req.DeskripsiPekerjaan == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Semua field wajib diisi dan tidak boleh kosong",
		})
	}

	// Panggil repository untuk simpan ke MongoDB
	newPekerjaan, err := s.repo.Create(ctx, &req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Gagal menyimpan data pekerjaan",
			"details": err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"message": "Data pekerjaan alumni berhasil ditambahkan",
		"data":    newPekerjaan,
	})
}

// @Summary Update pekerjaan
// @Description Update data pekerjaan (admin only)
// @Tags Pekerjaan_Alumni
// @Accept json
// @Produce json
// @Param id path string true "Pekerjaan ID"
// @Param pekerjaanRequest body models.UpdatePekerjaanRequest true "Update Pekerjaan Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security Bearer
// @Router /unair/pekerjaan-alumni/{id} [put]
func (s *PekerjaanService) Update(c *fiber.Ctx) error {
	ctx := context.Background()
	id := c.Params("id")

	var req models.UpdatePekerjaanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Body request tidak valid"})
	}

	if req.NamaPerusahaan == "" || req.PosisiJabatan == "" ||
		req.BidangIndustri == "" || req.LokasiKerja == "" || req.GajiRange == "" ||
		req.StatusPekerjaan == "" || req.DeskripsiPekerjaan == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Semua field wajib diisi"})
	}

	updated, err := s.repo.Update(ctx, id, &req)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return c.Status(404).JSON(fiber.Map{"error": "Pekerjaan tidak ditemukan"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "data": updated})
}

// @Summary Soft delete pekerjaan
// @Description Hapus data pekerjaan (soft delete)
// @Tags Pekerjaan_Alumni
// @Produce json
// @Param id path string true "Pekerjaan ID"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security Bearer
// @Router /unair/pekerjaan-alumni/{id} [delete]
func (s *PekerjaanService) SoftDelete(c *fiber.Ctx) error {
	ctx := context.Background()
	id := c.Params("id")
	role := c.Locals("role").(string)
	alumniID, _ := c.Locals("alumni_id").(string)

	if role == "admin" {
		if err := s.repo.SoftDeleteByID(ctx, id); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"success": true, "message": "Data pekerjaan berhasil dihapus oleh admin"})
	}

	// User hanya boleh hapus miliknya sendiri
	data, _ := s.repo.GetByID(ctx, id)
	if data == nil || data.AlumniID.Hex() != alumniID {
		return c.Status(403).JSON(fiber.Map{
			"error": "Kamu tidak bisa menghapus data milik orang lain",
		})
	}

	if err := s.repo.SoftDeleteByOwner(ctx, id, alumniID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Pekerjaanmu berhasil dihapus",
	})
}

// @Summary Restore deleted pekerjaan
// @Description Restore data pekerjaan yang telah dihapus
// @Tags Pekerjaan_Alumni
// @Produce json
// @Param id path string true "Pekerjaan ID"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security Bearer
// @Router /unair/pekerjaan-alumni/{id} [patch]
func (s *PekerjaanService) Restore(c *fiber.Ctx) error {
	ctx := context.Background()
	id := c.Params("id")
	role := c.Locals("role").(string)
	alumniID, _ := c.Locals("alumni_id").(string)

	var alumniPtr *string
	if alumniID != "" {
		alumniPtr = &alumniID
	}

	if err := s.repo.Restore(ctx, id, alumniPtr, role); err != nil {
		return c.Status(403).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Data berhasil direstore"})
}

// @Summary Get deleted pekerjaan (trash)
// @Description Get data pekerjaan yang telah dihapus (soft delete)
// @Tags Pekerjaan_Alumni
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security Bearer
// @Router /unair/pekerjaan-alumni/trash [get]
func (s *PekerjaanService) GetTrash(c *fiber.Ctx) error {
	ctx := context.Background()
	role := c.Locals("role").(string)

	if role == "admin" {
		data, err := s.repo.GetTrash(ctx)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"success": true, "data": data})
	}

	alumniID := c.Locals("alumni_id").(string)
	data, err := s.repo.GetTrashByOwner(ctx, alumniID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": data})
}

// @Summary Hard delete pekerjaan permanently
// @Description Hapus permanen data pekerjaan dari database
// @Tags Pekerjaan_Alumni
// @Produce json
// @Param id path string true "Pekerjaan ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security Bearer
// @Router /unair/pekerjaan-alumni/trash/delete/{id} [delete]
func (s *PekerjaanService) Delete(c *fiber.Ctx) error {
	ctx := context.Background()
	id := c.Params("id")
	role := c.Locals("role").(string)

	if role == "admin" {
		if err := s.repo.Delete(ctx, id, nil); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"success": true, "message": "Pekerjaan berhasil dihapus permanen oleh admin"})
	}

	alumniID := c.Locals("alumni_id").(string)
	if err := s.repo.Delete(ctx, id, &alumniID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Pekerjaanmu berhasil dihapus permanen"})
}

// @Summary Get pekerjaan with pagination
// @Description Get data pekerjaan dengan fitur pagination, search, dan sort
// @Tags Pekerjaan_Alumni
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search keyword"
// @Param sortBy query string false "Sort by field" default(created_at)
// @Param order query string false "Sort order" default(asc)
// @Success 200 {object} models.PekerjaanResponse
// @Failure 500 {object} map[string]interface{}
// @Security Bearer
// @Router /unair/pekerjaan-alumni [get]
func (s *PekerjaanService) GetPekerjaanService(c *fiber.Ctx) error {
	ctx := context.Background()

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	sortBy := c.Query("sortBy", "created_at")
	order := c.Query("order", "asc")
	search := c.Query("search", "")

	offset := int64((page - 1) * limit)

	data, err := s.repo.GetPekerjaanRepo(ctx, search, sortBy, order, int64(limit), offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	total, err := s.repo.CountPekerjaanRepo(ctx, search)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	response := models.PekerjaanResponse{
		Data: data,
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
