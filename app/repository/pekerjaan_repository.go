package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	models "crud-app/app/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PekerjaanRepository interface {
	GetAll(ctx context.Context) ([]models.Pekerjaan, error)
	GetByID(ctx context.Context, id string) (*models.Pekerjaan, error)
	GetByAlumniID(ctx context.Context, alumniID string) ([]models.Pekerjaan, error)
	Create(ctx context.Context, req *models.CreatePekerjaanRequest) (*models.Pekerjaan, error)
	Update(ctx context.Context, id string, req *models.UpdatePekerjaanRequest) (*models.Pekerjaan, error)
	SoftDeleteByID(ctx context.Context, id string) error
	SoftDeleteByOwner(ctx context.Context, id, alumniID string) error
	Restore(ctx context.Context, id string, alumniID *string, role string) error
	GetTrash(ctx context.Context) ([]models.Trash, error)
	GetTrashByOwner(ctx context.Context, alumniID string) ([]models.Trash, error)
	Delete(ctx context.Context, id string, alumniID *string) error
	GetPekerjaanRepo(ctx context.Context, search, sortBy, order string, limit, offset int64) ([]models.Pekerjaan, error)
	CountPekerjaanRepo(ctx context.Context, search string) (int64, error)
}

type pekerjaanRepository struct {
	collection      *mongo.Collection
	trashCollection *mongo.Collection
}

func NewPekerjaanRepository(database *mongo.Database) PekerjaanRepository {
	return &pekerjaanRepository{
		collection:      database.Collection("pekerjaan_alumni"),
		trashCollection: database.Collection("trash_pekerjaan"),
	}
}

// ========================== CREATE ==========================
func (r *pekerjaanRepository) Create(ctx context.Context, req *models.CreatePekerjaanRequest) (*models.Pekerjaan, error) {
	alumniObjID, err := primitive.ObjectIDFromHex(req.AlumniID)
	if err != nil {
		return nil, fmt.Errorf("alumni_id tidak valid: %v", err)
	}

	tglMulai, err := time.Parse(time.RFC3339, req.TanggalMulaiKerja)
	if err != nil {
		return nil, fmt.Errorf("format tanggal_mulai_kerja tidak valid")
	}

	var tglSelesaiPtr *time.Time
	if req.TanggalSelesaiKerja != "" {
		tglSelesai, err := time.Parse(time.RFC3339, req.TanggalSelesaiKerja)
		if err != nil {
			return nil, fmt.Errorf("format tanggal_selesai_kerja tidak valid")
		}
		tglSelesaiPtr = &tglSelesai
	}

	newPekerjaan := models.Pekerjaan{
		ID:                  primitive.NewObjectID(),
		AlumniID:            alumniObjID,
		NamaPerusahaan:      req.NamaPerusahaan,
		PosisiJabatan:       req.PosisiJabatan,
		BidangIndustri:      req.BidangIndustri,
		LokasiKerja:         req.LokasiKerja,
		GajiRange:           req.GajiRange,
		TanggalMulaiKerja:   tglMulai,
		TanggalSelesaiKerja: tglSelesaiPtr,
		StatusPekerjaan:     req.StatusPekerjaan,
		DeskripsiPekerjaan:  req.DeskripsiPekerjaan,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	_, err = r.collection.InsertOne(ctx, newPekerjaan)
	if err != nil {
		return nil, err
	}
	return &newPekerjaan, nil
}

// ========================== GET ALL ==========================
func (r *pekerjaanRepository) GetAll(ctx context.Context) ([]models.Pekerjaan, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.Pekerjaan
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// ========================== GET BY ID ==========================
func (r *pekerjaanRepository) GetByID(ctx context.Context, id string) (*models.Pekerjaan, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var pekerjaan models.Pekerjaan
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&pekerjaan)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &pekerjaan, err
}

// ========================== GET BY ALUMNI ==========================
func (r *pekerjaanRepository) GetByAlumniID(ctx context.Context, alumniID string) ([]models.Pekerjaan, error) {
	alumniObjID, err := primitive.ObjectIDFromHex(alumniID)
	if err != nil {
		return nil, err
	}

	cursor, err := r.collection.Find(ctx, bson.M{"alumni_id": alumniObjID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []models.Pekerjaan
	if err = cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// ========================== UPDATE ==========================
func (r *pekerjaanRepository) Update(ctx context.Context, id string, req *models.UpdatePekerjaanRequest) (*models.Pekerjaan, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	update := bson.M{
		"$set": bson.M{
			"nama_perusahaan":       req.NamaPerusahaan,
			"posisi_jabatan":        req.PosisiJabatan,
			"bidang_industri":       req.BidangIndustri,
			"lokasi_kerja":          req.LokasiKerja,
			"gaji_range":            req.GajiRange,
			"tanggal_mulai_kerja":   req.TanggalMulaiKerja,
			"tanggal_selesai_kerja": req.TanggalSelesaiKerja,
			"status_pekerjaan":      req.StatusPekerjaan,
			"deskripsi_pekerjaan":   req.DeskripsiPekerjaan,
			"updated_at":            time.Now(),
		},
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

// ========================== SOFT DELETE ==========================
func (r *pekerjaanRepository) SoftDeleteByID(ctx context.Context, id string) error {
	objID, _ := primitive.ObjectIDFromHex(id)

	var pekerjaan models.Pekerjaan
	if err := r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&pekerjaan); err != nil {
		return fmt.Errorf("data tidak ditemukan")
	}

	trash := models.Trash{
		ID:                  pekerjaan.ID,
		AlumniID:            pekerjaan.AlumniID,
		NamaPerusahaan:      pekerjaan.NamaPerusahaan,
		PosisiJabatan:       pekerjaan.PosisiJabatan,
		BidangIndustri:      pekerjaan.BidangIndustri,
		LokasiKerja:         pekerjaan.LokasiKerja,
		GajiRange:           pekerjaan.GajiRange,
		TanggalMulaiKerja:   pekerjaan.TanggalMulaiKerja,
		TanggalSelesaiKerja: pekerjaan.TanggalSelesaiKerja,
		StatusPekerjaan:     pekerjaan.StatusPekerjaan,
		DeskripsiPekerjaan:  pekerjaan.DeskripsiPekerjaan,
		IsDeleted:           true,
		CreatedAt:           pekerjaan.CreatedAt,
		UpdatedAt:           time.Now(),
	}

	if _, err := r.trashCollection.InsertOne(ctx, trash); err != nil {
		return err
	}

	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (r *pekerjaanRepository) SoftDeleteByOwner(ctx context.Context, id, alumniID string) error {
	objID, _ := primitive.ObjectIDFromHex(id)
	alumniObj, _ := primitive.ObjectIDFromHex(alumniID)

	var pekerjaan models.Pekerjaan
	if err := r.collection.FindOne(ctx, bson.M{"_id": objID, "alumni_id": alumniObj}).Decode(&pekerjaan); err != nil {
		return fmt.Errorf("data tidak ditemukan atau bukan milik user ini")
	}

	trash := models.Trash{
		ID:                  pekerjaan.ID,
		AlumniID:            pekerjaan.AlumniID,
		NamaPerusahaan:      pekerjaan.NamaPerusahaan,
		PosisiJabatan:       pekerjaan.PosisiJabatan,
		BidangIndustri:      pekerjaan.BidangIndustri,
		LokasiKerja:         pekerjaan.LokasiKerja,
		GajiRange:           pekerjaan.GajiRange,
		TanggalMulaiKerja:   pekerjaan.TanggalMulaiKerja,
		TanggalSelesaiKerja: pekerjaan.TanggalSelesaiKerja,
		StatusPekerjaan:     pekerjaan.StatusPekerjaan,
		DeskripsiPekerjaan:  pekerjaan.DeskripsiPekerjaan,
		IsDeleted:           true,
		CreatedAt:           pekerjaan.CreatedAt,
		UpdatedAt:           time.Now(),
	}

	if _, err := r.trashCollection.InsertOne(ctx, trash); err != nil {
		return err
	}

	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": objID, "alumni_id": alumniObj})
	return err
}

// ========================== RESTORE ==========================
func (r *pekerjaanRepository) Restore(ctx context.Context, id string, alumniID *string, role string) error {
	objID, _ := primitive.ObjectIDFromHex(id)

	filter := bson.M{"_id": objID}

	// Alumni hanya boleh restore miliknya sendiri
	if role != "admin" && alumniID != nil {
		alumniObj, _ := primitive.ObjectIDFromHex(*alumniID)
		filter["alumni_id"] = alumniObj
	}

	var trash models.Trash
	if err := r.trashCollection.FindOne(ctx, filter).Decode(&trash); err != nil {
		return fmt.Errorf("data tidak ditemukan di trash atau kamu tidak punya akses")
	}

	pekerjaan := models.Pekerjaan{
		ID:                  trash.ID,
		AlumniID:            trash.AlumniID,
		NamaPerusahaan:      trash.NamaPerusahaan,
		PosisiJabatan:       trash.PosisiJabatan,
		BidangIndustri:      trash.BidangIndustri,
		LokasiKerja:         trash.LokasiKerja,
		GajiRange:           trash.GajiRange,
		TanggalMulaiKerja:   trash.TanggalMulaiKerja,
		TanggalSelesaiKerja: trash.TanggalSelesaiKerja,
		StatusPekerjaan:     trash.StatusPekerjaan,
		DeskripsiPekerjaan:  trash.DeskripsiPekerjaan,
		CreatedAt:           trash.CreatedAt,
		UpdatedAt:           time.Now(),
	}

	if _, err := r.collection.InsertOne(ctx, pekerjaan); err != nil {
		return err
	}

	_, err := r.trashCollection.DeleteOne(ctx, filter)
	return err
}

// ========================== GET TRASH ==========================
func (r *pekerjaanRepository) GetTrash(ctx context.Context) ([]models.Trash, error) {
	cursor, err := r.trashCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []models.Trash
	if err = cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *pekerjaanRepository) GetTrashByOwner(ctx context.Context, alumniID string) ([]models.Trash, error) {
	alumniObj, err := primitive.ObjectIDFromHex(alumniID)
	if err != nil {
		return nil, fmt.Errorf("alumni_id tidak valid")
	}

	cursor, err := r.trashCollection.Find(ctx, bson.M{"alumni_id": alumniObj})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []models.Trash
	if err = cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// ========================== DELETE (HARD) ==========================
func (r *pekerjaanRepository) Delete(ctx context.Context, id string, alumniID *string) error {
	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": objID}
	if alumniID != nil {
		alumniObj, _ := primitive.ObjectIDFromHex(*alumniID)
		filter["alumni_id"] = alumniObj
	}
	_, err := r.trashCollection.DeleteOne(ctx, filter)
	return err
}

// ========================== SEARCH, SORT, PAGINATION ==========================
func (r *pekerjaanRepository) GetPekerjaanRepo(ctx context.Context, search, sortBy, order string, limit, offset int64) ([]models.Pekerjaan, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"nama_perusahaan": bson.M{"$regex": search, "$options": "i"}},
			{"posisi_jabatan": bson.M{"$regex": search, "$options": "i"}},
			{"bidang_industri": bson.M{"$regex": search, "$options": "i"}},
			{"lokasi_kerja": bson.M{"$regex": search, "$options": "i"}},
		},
	}

	sortOrder := 1
	if order == "desc" {
		sortOrder = -1
	}

	opts := options.Find().
		SetSort(bson.D{{Key: sortBy, Value: sortOrder}}).
		SetLimit(limit).
		SetSkip(offset)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		log.Println("Query error:", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var pekerjaan []models.Pekerjaan
	if err = cursor.All(ctx, &pekerjaan); err != nil {
		return nil, err
	}
	return pekerjaan, nil
}

func (r *pekerjaanRepository) CountPekerjaanRepo(ctx context.Context, search string) (int64, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"nama_perusahaan": bson.M{"$regex": search, "$options": "i"}},
			{"posisi_jabatan": bson.M{"$regex": search, "$options": "i"}},
			{"bidang_industri": bson.M{"$regex": search, "$options": "i"}},
			{"lokasi_kerja": bson.M{"$regex": search, "$options": "i"}},
		},
	}
	count, err := r.collection.CountDocuments(ctx, filter)
	return count, err
}
