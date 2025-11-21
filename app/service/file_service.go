package service

import (
	"fmt"
	"os"
	"path/filepath"

	models "crud-app/app/model"
	"crud-app/app/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileService interface {
	UploadFoto(c *fiber.Ctx) error
	UploadSertifikat(c *fiber.Ctx) error
	GetAllFiles(c *fiber.Ctx) error
	GetFileByID(c *fiber.Ctx) error
	DeleteFile(c *fiber.Ctx) error
}

type fileService struct {
	repo       repository.FileRepository
	uploadPath string
}

func NewFileService(repo repository.FileRepository, uploadPath string) FileService {
	return &fileService{
		repo:       repo,
		uploadPath: uploadPath,
	}
}

// @Summary Upload foto (profile picture)
// @Description Upload file foto dengan maksimal 1 MB
// @Accept mpfd
// @Tags Files
// @Produce json
// @Param file formData file true "Image file (JPEG/PNG)"
// @Param user_id formData string false "User ID (untuk admin upload)"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 413 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security Bearer
// @Router /files/upload/foto [post]
func (s *fileService) UploadFoto(c *fiber.Ctx) error {

	allowedTypes := []string{"image/jpeg", "image/jpg", "image/png"}
	return s.uploadHandler(c, allowedTypes, 1)
}

// @Summary Upload sertifikat (PDF)
// @Description Upload file sertifikat dalam format PDF dengan maksimal 2 MB
// @Tags Files
// @Accept mpfd
// @Produce json
// @Param file formData file true "PDF file"
// @Param user_id formData string false "User ID (untuk admin upload)"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 413 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security Bearer
// @Router /files/upload/sertifikat [post]
func (s *fileService) UploadSertifikat(c *fiber.Ctx) error {

	allowedTypes := []string{"application/pdf"}
	return s.uploadHandler(c, allowedTypes, 2)
}

// @Summary Internal upload handler
// @Description Handler internal untuk upload file
// @Hidden
// @Tags Files
// @Security Bearer
func (s *fileService) uploadHandler(c *fiber.Ctx, allowedTypes []string, maxSizeMB int64) error {

	roleVal := c.Locals("role")
	userIDVal := c.Locals("user_id")

	if roleVal == nil || userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Token tidak valid atau role tidak dikenali",
		})
	}

	role, _ := roleVal.(string)
	userIDStr := fmt.Sprintf("%v", userIDVal)
	var targetUserID primitive.ObjectID
	var err error

	// 🔹 Role-based access logic
	switch role {
	case "user", "alumni":
		formUserID := c.FormValue("user_id")
		if formUserID != "" && formUserID != userIDStr {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Kamu tidak boleh upload file untuk user lain",
			})
		}
		targetUserID, _ = primitive.ObjectIDFromHex(userIDStr)

	case "admin":
		targetUserIDHex := c.FormValue("user_id")
		if targetUserIDHex == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "user_id is required for admin upload",
			})
		}
		targetUserID, err = primitive.ObjectIDFromHex(targetUserIDHex)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid user_id format",
			})
		}

	default:
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Role '%v' tidak diizinkan upload", role),
		})
	}

	// 🔹 Ambil file dari form
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "No file uploaded",
		})
	}

	// 🔹 Validasi tipe file
	contentType := fileHeader.Header.Get("Content-Type")
	allowed := false
	for _, t := range allowedTypes {
		if t == contentType {
			allowed = true
			break
		}
	}
	if !allowed {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Invalid file type. Allowed: %v", allowedTypes),
		})
	}

	// 🔹 Validasi ukuran file
	maxSize := maxSizeMB * 1024 * 1024
	if fileHeader.Size > maxSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("File too large (max %d MB)", maxSizeMB),
		})
	}

	// 🔹 Simpan file ke folder
	ext := filepath.Ext(fileHeader.Filename)
	newFileName := uuid.New().String() + ext
	filePath := filepath.Join(s.uploadPath, newFileName)
	os.MkdirAll(s.uploadPath, os.ModePerm)

	if err := c.SaveFile(fileHeader, filePath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to save file",
		})
	}

	// 🔹 Simpan metadata ke database
	fileModel := &models.File{
		UserID:       &targetUserID,
		FileName:     newFileName,
		OriginalName: fileHeader.Filename,
		FilePath:     filePath,
		FileSize:     fileHeader.Size,
		FileType:     contentType,
	}

	if err := s.repo.Create(fileModel); err != nil {
		os.Remove(filePath)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to save file metadata",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "File uploaded successfully",
		"data":    s.toFileResponse(fileModel),
	})
}

// @Summary Get all files
// @Description Get daftar semua file (admin melihat semua, user hanya miliknya)
// @Tags Files
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security Bearer
// @Router /files [get]
func (s *fileService) GetAllFiles(c *fiber.Ctx) error {

	role := c.Locals("role").(string)
	userIDStr := c.Locals("user_id").(string)

	var files []models.File
	var err error

	if role == "admin" {
		files, err = s.repo.FindAll()
	} else {
		userID, _ := primitive.ObjectIDFromHex(userIDStr)
		files, err = s.repo.FindByUserID(userID)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get files",
		})
	}

	var responses []models.FileResponse
	for _, f := range files {
		responses = append(responses, *s.toFileResponse(&f))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    responses,
	})
}

// @Summary Get file by ID
// @Description Get detail file berdasarkan ID
// @Tags Files
// @Produce json
// @Param id path string true "File ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security Bearer
// @Router /files/{id} [get]
func (s *fileService) GetFileByID(c *fiber.Ctx) error {

	id := c.Params("id")
	file, err := s.repo.FindByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "File not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    s.toFileResponse(file),
	})
}

// @Summary Delete file
// @Description Hapus file dari storage dan database
// @Tags Files
// @Produce json
// @Param id path string true "File ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security Bearer
// @Router /files/{id} [delete]
func (s *fileService) DeleteFile(c *fiber.Ctx) error {

	id := c.Params("id")
	file, err := s.repo.FindByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "File not found",
		})
	}

	if err := os.Remove(file.FilePath); err != nil {
		fmt.Println("Warning: failed to remove file from storage:", err)
	}

	if err := s.repo.Delete(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete file",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "File deleted successfully",
	})
}

func (s *fileService) toFileResponse(file *models.File) *models.FileResponse {
	return &models.FileResponse{
		ID:           file.ID.Hex(),
		FileName:     file.FileName,
		OriginalName: file.OriginalName,
		FilePath:     file.FilePath,
		FileSize:     file.FileSize,
		FileType:     file.FileType,
		UploadedAt:   file.UploadedAt,
	}
}
