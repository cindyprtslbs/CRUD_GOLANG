package route

import (
	"crud-app/app/repository"
	"crud-app/app/service"
	"crud-app/middleware"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetupRoutes godoc
// @Summary Setup semua routes untuk aplikasi CRUD
// @Description Mengatur semua endpoint untuk autentikasi, alumni, pekerjaan, dan upload file
// @Tags Base
// @Router /api [get]
func SetupRoutes(app *fiber.App, db *mongo.Database) {
	// -------------------------
	// Base groups
	// -------------------------
	api := app.Group("/api")
	unair := app.Group("/unair")

	// =========================
	// AUTH ROUTES
	// =========================
	authRepo := repository.NewAuthRepository(db)
	authService := service.NewAuthService(authRepo)

	// @Summary Login user
	// @Description Autentikasi user dan mengembalikan token JWT
	// @Tags Auth
	// @Accept json
	// @Produce json
	// @Param body body map[string]string true "Email dan Password"
	// @Success 200 {object} map[string]interface{}
	// @Failure 401 {object} map[string]interface{}
	// @Router /api/login [post]
	api.Post("/login", authService.Login)

	// Protected route (harus login)
	protected := api.Group("", middleware.AuthRequired())

	// @Summary Mendapatkan profil user yang sedang login
	// @Description Mengambil data profil user berdasarkan token JWT
	// @Tags Auth
	// @Produce json
	// @Success 200 {object} map[string]interface{}
	// @Failure 401 {object} map[string]interface{}
	// @Router /api/profile [get]
	protected.Get("/profile", authService.GetProfile)

	// =========================
	// ALUMNI ROUTES
	// =========================
	alumniRepo := repository.NewAlumniRepository(db)
	alumniService := service.NewAlumniService(alumniRepo)

	alumni := unair.Group("/alumni")

	// @Summary Mendapatkan daftar alumni dengan pagination dan filter
	// @Description Mendapatkan daftar alumni berdasarkan pencarian, sorting, dan pagination
	// @Tags Alumni
	// @Produce json
	// @Param page query int false "Nomor halaman"
	// @Param limit query int false "Jumlah data per halaman"
	// @Param sortBy query string false "Kolom untuk sorting"
	// @Param order query string false "Urutan sorting (asc/desc)"
	// @Param search query string false "Kata kunci pencarian"
	// @Success 200 {object} map[string]interface{}
	// @Failure 500 {object} map[string]interface{}
	// @Router /unair/alumni [get]
	alumni.Get("/", alumniService.GetAlumniService)

	// @Summary Mendapatkan alumni tanpa pekerjaan
	// @Tags Alumni
	// @Produce json
	// @Success 200 {object} map[string]interface{}
	// @Router /unair/alumni/without-pekerjaan [get]
	alumni.Get("/without-pekerjaan", alumniService.GetWithoutPekerjaan)

	// @Summary Mendapatkan data alumni berdasarkan ID
	// @Tags Alumni
	// @Param id path string true "ID Alumni"
	// @Produce json
	// @Success 200 {object} map[string]interface{}
	// @Failure 404 {object} map[string]interface{}
	// @Router /unair/alumni/{id} [get]
	alumni.Get("/:id", alumniService.GetByID)

	// @Summary Menambah data alumni baru
	// @Tags Alumni
	// @Accept json
	// @Produce json
	// @Param body body map[string]interface{} true "Data Alumni Baru"
	// @Success 201 {object} map[string]interface{}
	// @Failure 400 {object} map[string]interface{}
	// @Router /unair/alumni [post]
	alumni.Post("/", middleware.AuthRequired(), middleware.AdminOnly(), alumniService.Create)

	// @Summary Mengupdate data alumni
	// @Tags Alumni
	// @Accept json
	// @Produce json
	// @Param id path string true "ID Alumni"
	// @Param body body map[string]interface{} true "Data Alumni yang Diupdate"
	// @Success 200 {object} map[string]interface{}
	// @Failure 404 {object} map[string]interface{}
	// @Router /unair/alumni/{id} [put]
	alumni.Put("/:id", middleware.AuthRequired(), middleware.AdminOnly(), alumniService.Update)

	// @Summary Menghapus alumni (soft delete)
	// @Tags Alumni
	// @Param id path string true "ID Alumni"
	// @Success 200 {object} map[string]interface{}
	// @Failure 404 {object} map[string]interface{}
	// @Router /unair/alumni/{id} [delete]
	alumni.Delete("/:id", middleware.AuthRequired(), alumniService.SoftDelete)

	// @Summary Merestore alumni yang dihapus
	// @Tags Alumni
	// @Param id path string true "ID Alumni"
	// @Success 200 {object} map[string]interface{}
	// @Router /unair/alumni/{id} [patch]
	alumni.Patch("/:id", middleware.AuthRequired(), alumniService.Restore)

	// =========================
	// PEKERJAAN ALUMNI ROUTES
	// =========================
	pekerjaanRepo := repository.NewPekerjaanRepository(db)
	pekerjaanService := service.NewPekerjaanService(pekerjaanRepo)

	pekerjaan := unair.Group("/pekerjaan-alumni")

	// @Summary Mendapatkan daftar pekerjaan alumni
	// @Tags Pekerjaan
	// @Produce json
	// @Success 200 {object} map[string]interface{}
	// @Router /unair/pekerjaan-alumni [get]
	pekerjaan.Get("/", pekerjaanService.GetPekerjaanService)

	// @Summary Mendapatkan daftar pekerjaan alumni yang dihapus (trash)
	// @Tags Pekerjaan
	// @Produce json
	// @Success 200 {object} map[string]interface{}
	// @Router /unair/pekerjaan-alumni/trash [get]
	pekerjaan.Get("/trash", middleware.AuthRequired(), pekerjaanService.GetTrash)

	// @Summary Mendapatkan data pekerjaan berdasarkan ID
	// @Tags Pekerjaan
	// @Param id path string true "ID Pekerjaan"
	// @Produce json
	// @Success 200 {object} map[string]interface{}
	// @Failure 404 {object} map[string]interface{}
	// @Router /unair/pekerjaan-alumni/{id} [get]
	pekerjaan.Get("/:id", middleware.AuthRequired(), pekerjaanService.GetByID)

	// @Summary Mendapatkan pekerjaan berdasarkan ID alumni
	// @Tags Pekerjaan
	// @Param alumni_id path string true "ID Alumni"
	// @Produce json
	// @Success 200 {object} map[string]interface{}
	// @Router /unair/pekerjaan-alumni/alumni/{alumni_id} [get]
	pekerjaan.Get("/alumni/:alumni_id", middleware.AuthRequired(), middleware.AdminOnly(), pekerjaanService.GetByAlumniID)

	// @Summary Menambah data pekerjaan baru
	// @Tags Pekerjaan
	// @Accept json
	// @Produce json
	// @Param body body map[string]interface{} true "Data Pekerjaan Baru"
	// @Success 201 {object} map[string]interface{}
	// @Router /unair/pekerjaan-alumni [post]
	pekerjaan.Post("/", middleware.AuthRequired(), middleware.AdminOnly(), pekerjaanService.Create)

	// @Summary Mengupdate data pekerjaan
	// @Tags Pekerjaan
	// @Param id path string true "ID Pekerjaan"
	// @Param body body map[string]interface{} true "Data Pekerjaan"
	// @Success 200 {object} map[string]interface{}
	// @Router /unair/pekerjaan-alumni/{id} [put]
	pekerjaan.Put("/:id", middleware.AuthRequired(), middleware.AdminOnly(), pekerjaanService.Update)

	// @Summary Menghapus data pekerjaan (soft delete)
	// @Tags Pekerjaan
	// @Param id path string true "ID Pekerjaan"
	// @Success 200 {object} map[string]interface{}
	// @Router /unair/pekerjaan-alumni/{id} [delete]
	pekerjaan.Delete("/:id", middleware.AuthRequired(), pekerjaanService.SoftDelete)

	// @Summary Merestore data pekerjaan
	// @Tags Pekerjaan
	// @Param id path string true "ID Pekerjaan"
	// @Success 200 {object} map[string]interface{}
	// @Router /unair/pekerjaan-alumni/{id} [patch]
	pekerjaan.Patch("/:id", middleware.AuthRequired(), pekerjaanService.Restore)

	// @Summary Merestore pekerjaan dari trash
	// @Tags Pekerjaan
	// @Param id path string true "ID Pekerjaan"
	// @Router /unair/pekerjaan-alumni/restore/{id} [put]
	pekerjaan.Put("/restore/:id", middleware.AuthRequired(), pekerjaanService.Restore)

	// @Summary Menghapus permanen pekerjaan dari trash
	// @Tags Pekerjaan
	// @Param id path string true "ID Pekerjaan"
	// @Router /unair/pekerjaan-alumni/trash/delete/{id} [delete]
	pekerjaan.Delete("/trash/delete/:id", middleware.AuthRequired(), pekerjaanService.Delete)

	// =========================
	// UPLOAD FILES ROUTES
	// =========================
	files := app.Group("/files")

	fileRepo := repository.NewFileRepository(db)
	uploadPath := "./uploads"
	fileService := service.NewFileService(fileRepo, uploadPath)

	// @Summary Upload foto alumni
	// @Tags Upload
	// @Accept multipart/form-data
	// @Param file formData file true "File foto"
	// @Success 201 {object} map[string]interface{}
	// @Router /files/upload/foto [post]
	files.Post("/upload/foto", middleware.AuthRequired(), fileService.UploadFoto)

	// @Summary Upload sertifikat alumni
	// @Tags Upload
	// @Accept multipart/form-data
	// @Param file formData file true "File sertifikat"
	// @Success 201 {object} map[string]interface{}
	// @Router /files/upload/sertifikat [post]
	files.Post("/upload/sertifikat", middleware.AuthRequired(), fileService.UploadSertifikat)

	// @Summary Mendapatkan semua file yang diupload
	// @Tags Upload
	// @Success 200 {object} map[string]interface{}
	// @Router /files [get]
	files.Get("/", middleware.AuthRequired(), fileService.GetAllFiles)

	// @Summary Mendapatkan file berdasarkan ID
	// @Tags Upload
	// @Param id path string true "ID File"
	// @Success 200 {object} map[string]interface{}
	// @Router /files/{id} [get]
	files.Get("/:id", middleware.AuthRequired(), fileService.GetFileByID)

	// @Summary Menghapus file berdasarkan ID
	// @Tags Upload
	// @Param id path string true "ID File"
	// @Success 200 {object} map[string]interface{}
	// @Router /files/{id} [delete]
	files.Delete("/:id", middleware.AuthRequired(), fileService.DeleteFile)
}
