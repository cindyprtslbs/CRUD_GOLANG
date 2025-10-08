package route

import (
	"crud-app/app/repository"
	"crud-app/app/service"
	"crud-app/middleware"
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, db *sql.DB) {
	// -------------------------
	// Base groups
	// -------------------------
	api := app.Group("/api")
	unair := app.Group("/unair")

	// -------------------------
	// Auth routes
	// -------------------------
	authRepo := repository.NewAuthRepository(db)
	authService := service.NewAuthService(authRepo)

	api.Post("/login", authService.Login)

	// Protected (user sudah login)
	protected := api.Group("", middleware.AuthRequired())
	protected.Get("/profile", authService.GetProfile)

	// -------------------------
	// Alumni routes
	// -------------------------
	alumniRepo := repository.NewAlumniRepository(db)
	alumniService := service.NewAlumniService(alumniRepo)

	alumni := unair.Group("/alumni")
	alumni.Get("/", alumniService.GetAlumniService)
	alumni.Get("/without-pekerjaan", alumniService.GetWithoutPekerjaan)
	alumni.Get("/:id", alumniService.GetByID)

	alumni.Post("/", middleware.AuthRequired(), middleware.AdminOnly(), alumniService.Create)
	alumni.Put("/:id", middleware.AuthRequired(), middleware.AdminOnly(), alumniService.Update)
	alumni.Delete("/:id", middleware.AuthRequired(), alumniService.SoftDelete)
	alumni.Patch("/:id", middleware.AuthRequired(), alumniService.Restore)

	// -------------------------
	// Pekerjaan Alumni routes
	// -------------------------
	pekerjaanRepo := repository.NewPekerjaanRepository(db)
	pekerjaanService := service.NewPekerjaanService(pekerjaanRepo)

	pekerjaan := unair.Group("/pekerjaan-alumni")
	pekerjaan.Get("/", pekerjaanService.GetPekerjaanService)
	pekerjaan.Get("/pekerja-lebih-setahun", pekerjaanService.GetBekerjalebih1Tahun)
	pekerjaan.Get("/trash", middleware.AuthRequired(), pekerjaanService.GetTrash)
	pekerjaan.Get("/:id", middleware.AuthRequired(), pekerjaanService.GetByID)
	pekerjaan.Get("/alumni/:alumni_id", middleware.AuthRequired(), middleware.AdminOnly(), pekerjaanService.GetByAlumniID)

	pekerjaan.Post("/", middleware.AuthRequired(), middleware.AdminOnly(), pekerjaanService.Create)
	pekerjaan.Put("/:id", middleware.AuthRequired(), middleware.AdminOnly(), pekerjaanService.Update)
	pekerjaan.Delete("/:id", middleware.AuthRequired(), pekerjaanService.SoftDelete)
	pekerjaan.Patch("/:id", middleware.AuthRequired(), pekerjaanService.Restore)

	pekerjaan.Put("/restore/:id", middleware.AuthRequired(), pekerjaanService.Restore)
	pekerjaan.Delete("/trash/delete/:id", middleware.AuthRequired(), pekerjaanService.Delete)

	// endpoint pemanggilan /trash harus diletakkan di
}
