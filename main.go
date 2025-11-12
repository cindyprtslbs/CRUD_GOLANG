package main

import (
	"log"

	"crud-app/config"
	"crud-app/database"
	"crud-app/route"

	_ "crud-app/docs"

	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// @title		CRUD APPLICATION
// @version		1.0
// @description API sederhana untuk operasi CRUD menggunakan Fiber dan MongoDB
// @host 		localhost:3000
// @BasePath 	/
// @schemes 	http
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description "JWT Token dengan prefix 'Bearer '"

// @Summary Login user
// @Description Autentikasi user dan mengembalikan token JWT
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body map[string]string true "Email dan Password"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/login [post]
func SwaggerLogin(c *fiber.Ctx) error { return nil }

// @Summary Mendapatkan profil user yang sedang login
// @Description Mengambil data profil user berdasarkan token JWT
// @Tags Auth
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/profile [get]
func SwaggerProfile(c *fiber.Ctx) error { return nil }

// @Summary Mendapatkan daftar alumni dengan pagination dan filter
// @Tags Alumni
// @Produce json
// @Param page query int false "Nomor halaman" default(1)
// @Param limit query int false "Jumlah data per halaman" default(10)
// @Param sortBy query string false "Kolom untuk sorting"
// @Param order query string false "Urutan sorting (asc/desc)"
// @Param search query string false "Kata kunci pencarian"
// @Success 200 {object} map[string]interface{}
// @Router /unair/alumni [get]
func SwaggerGetAlumni(c *fiber.Ctx) error { return nil }

// @Summary Mendapatkan alumni tanpa pekerjaan
// @Tags Alumni
// @Produce json
// @Router /unair/alumni/without-pekerjaan [get]
func SwaggerGetWithoutPekerjaan(c *fiber.Ctx) error { return nil }

// @Summary Mendapatkan data alumni berdasarkan ID
// @Tags Alumni
// @Param id path string true "ID Alumni"
// @Produce json
// @Router /unair/alumni/{id} [get]
func SwaggerGetAlumniByID(c *fiber.Ctx) error { return nil }

// @Summary Menambah data alumni baru
// @Tags Alumni
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body map[string]interface{} true "Data Alumni Baru"
// @Router /unair/alumni [post]
func SwaggerCreateAlumni(c *fiber.Ctx) error { return nil }

// @Summary Mengupdate data alumni
// @Tags Alumni
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "ID Alumni"
// @Param body body map[string]interface{} true "Data Alumni yang Diupdate"
// @Router /unair/alumni/{id} [put]
func SwaggerUpdateAlumni(c *fiber.Ctx) error { return nil }

// @Summary Menghapus alumni (soft delete)
// @Tags Alumni
// @Param id path string true "ID Alumni"
// @Security ApiKeyAuth
// @Router /unair/alumni/{id} [delete]
func SwaggerSoftDeleteAlumni(c *fiber.Ctx) error { return nil }

// @Summary Merestore alumni yang dihapus
// @Tags Alumni
// @Param id path string true "ID Alumni"
// @Security ApiKeyAuth
// @Router /unair/alumni/{id} [patch]
func SwaggerRestoreAlumni(c *fiber.Ctx) error { return nil }

// @Summary Mendapatkan daftar pekerjaan alumni
// @Tags Pekerjaan
// @Produce json
// @Router /unair/pekerjaan-alumni [get]
func SwaggerGetPekerjaan(c *fiber.Ctx) error { return nil }

// @Summary Mendapatkan daftar pekerjaan alumni yang dihapus (trash)
// @Tags Pekerjaan
// @Produce json
// @Security ApiKeyAuth
// @Router /unair/pekerjaan-alumni/trash [get]
func SwaggerGetPekerjaanTrash(c *fiber.Ctx) error { return nil }

// @Summary Mendapatkan data pekerjaan berdasarkan ID
// @Tags Pekerjaan
// @Param id path string true "ID Pekerjaan"
// @Produce json
// @Security ApiKeyAuth
// @Router /unair/pekerjaan-alumni/{id} [get]
func SwaggerGetPekerjaanByID(c *fiber.Ctx) error { return nil }

// @Summary Mendapatkan pekerjaan berdasarkan ID alumni
// @Tags Pekerjaan
// @Param alumni_id path string true "ID Alumni"
// @Produce json
// @Security ApiKeyAuth
// @Router /unair/pekerjaan-alumni/alumni/{alumni_id} [get]
func SwaggerGetPekerjaanByAlumniID(c *fiber.Ctx) error { return nil }

// @Summary Menambah data pekerjaan baru
// @Tags Pekerjaan
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body map[string]interface{} true "Data Pekerjaan Baru"
// @Router /unair/pekerjaan-alumni [post]
func SwaggerCreatePekerjaan(c *fiber.Ctx) error { return nil }

// @Summary Mengupdate data pekerjaan
// @Tags Pekerjaan
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "ID Pekerjaan"
// @Param body body map[string]interface{} true "Data Pekerjaan"
// @Router /unair/pekerjaan-alumni/{id} [put]
func SwaggerUpdatePekerjaan(c *fiber.Ctx) error { return nil }

// @Summary Menghapus data pekerjaan (soft delete)
// @Tags Pekerjaan
// @Param id path string true "ID Pekerjaan"
// @Security ApiKeyAuth
// @Router /unair/pekerjaan-alumni/{id} [delete]
func SwaggerSoftDeletePekerjaan(c *fiber.Ctx) error { return nil }

// @Summary Merestore data pekerjaan
// @Tags Pekerjaan
// @Param id path string true "ID Pekerjaan"
// @Security ApiKeyAuth
// @Router /unair/pekerjaan-alumni/{id} [patch]
func SwaggerRestorePekerjaan(c *fiber.Ctx) error { return nil }

// @Summary Merestore pekerjaan dari trash
// @Tags Pekerjaan
// @Param id path string true "ID Pekerjaan"
// @Security ApiKeyAuth
// @Router /unair/pekerjaan-alumni/restore/{id} [put]
func SwaggerRestorePekerjaanFromTrash(c *fiber.Ctx) error { return nil }

// @Summary Menghapus permanen pekerjaan dari trash
// @Tags Pekerjaan
// @Param id path string true "ID Pekerjaan"
// @Security ApiKeyAuth
// @Router /unair/pekerjaan-alumni/trash/delete/{id} [delete]
func SwaggerDeletePekerjaanPermanent(c *fiber.Ctx) error { return nil }

// @Summary Upload foto alumni
// @Tags Upload
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param file formData file true "File foto"
// @Router /files/upload/foto [post]
func SwaggerUploadFoto(c *fiber.Ctx) error { return nil }

// @Summary Upload sertifikat alumni
// @Tags Upload
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param file formData file true "File sertifikat"
// @Router /files/upload/sertifikat [post]
func SwaggerUploadSertifikat(c *fiber.Ctx) error { return nil }

// @Summary Mendapatkan semua file yang diupload
// @Tags Upload
// @Produce json
// @Security ApiKeyAuth
// @Router /files [get]
func SwaggerGetAllFiles(c *fiber.Ctx) error { return nil }

// @Summary Mendapatkan file berdasarkan ID
// @Tags Upload
// @Param id path string true "ID File"
// @Produce json
// @Security ApiKeyAuth
// @Router /files/{id} [get]
func SwaggerGetFileByID(c *fiber.Ctx) error { return nil }

// @Summary Menghapus file berdasarkan ID
// @Tags Upload
// @Param id path string true "ID File"
// @Produce json
// @Security ApiKeyAuth
// @Router /files/{id} [delete]
func SwaggerDeleteFile(c *fiber.Ctx) error { return nil }

func main() {
	config.LoadEnv()
	config.InitLogger()

	db := database.ConnectMongo()

	app := config.NewApp()

	route.SetupRoutes(app, db)

	// Setup Swagger documentation
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Redirect root to Swagger UI
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/swagger/index.html", 301)
	})

	// Jalankan server
	port := config.GetEnv("APP_PORT", "3000")
	config.Logger.Println("Server running at http://localhost:" + port)
	config.Logger.Println("Swagger UI available at http://localhost:" + port + "/swagger/index.html")
	log.Fatal(app.Listen(":" + port))
}