package repository

import (
	"context"
	"time"
	"log"
	"fmt"

	models "crud-app/app/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ================= INTERFACE =================
type AlumniRepository interface {
	GetAll(ctx context.Context) ([]models.Alumni, error)
	GetByID(ctx context.Context, id string) (*models.Alumni, error)
	Create(ctx context.Context, req *models.CreateAlumniRequest) (*models.Alumni, error)
	Update(ctx context.Context, id string, req *models.UpdateAlumniRequest) (*models.Alumni, error)
	SoftDelete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	GetWithoutPekerjaan(ctx context.Context) ([]models.Alumni, error)
	CountWithoutPekerjaan(ctx context.Context) (int, error)
	GetAlumniRepo(ctx context.Context, search, sortBy, order string, limit, offset int64) ([]models.Alumni, error)
	CountAlumniRepo(ctx context.Context, search string) (int64, error)
}

// ================= STRUCT =================
type alumniRepository struct {
	collection *mongo.Collection
}

// ================= CONSTRUCTOR =================
func NewAlumniRepository(database *mongo.Database) AlumniRepository {
	return &alumniRepository{
		collection: database.Collection("alumni"),
	}
}

// ================= CREATE =================
func (r *alumniRepository) Create(ctx context.Context, req *models.CreateAlumniRequest) (*models.Alumni, error) {
	var userObjID *primitive.ObjectID
	if req.UserID != "" {
		id, err := primitive.ObjectIDFromHex(req.UserID)
		if err != nil {
			return nil, fmt.Errorf("user_id tidak valid: %v", err)
		}
		userObjID = &id
	}

	alumni := models.Alumni{
		ID:         primitive.NewObjectID(),
		UserID:     userObjID,
		NIM:        req.NIM,
		Nama:       req.Nama,
		Jurusan:    req.Jurusan,
		Angkatan:   req.Angkatan,
		TahunLulus: req.TahunLulus,
		Email:      req.Email,
		NoTelepon:  req.NoTelepon,
		Alamat:     req.Alamat,
		IsDeleted:  false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	_, err := r.collection.InsertOne(ctx, alumni)
	if err != nil {
		return nil, err
	}
	return &alumni, nil
}


// ================= GET ALL =================
func (r *alumniRepository) GetAll(ctx context.Context) ([]models.Alumni, error) {
	filter := bson.M{"is_deleted": false}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []models.Alumni
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// ================= GET BY ID =================
func (r *alumniRepository) GetByID(ctx context.Context, id string) (*models.Alumni, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var alumni models.Alumni
	err = r.collection.FindOne(ctx, bson.M{"_id": objID, "is_deleted": false}).Decode(&alumni)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &alumni, err
}

// ================= UPDATE =================
func (r *alumniRepository) Update(ctx context.Context, id string, req *models.UpdateAlumniRequest) (*models.Alumni, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	update := bson.M{
		"$set": bson.M{
			"user_id":      req.UserID,
			"nim":          req.NIM,
			"nama":         req.Nama,
			"jurusan":      req.Jurusan,
			"angkatan":     req.Angkatan,
			"tahun_lulus":  req.TahunLulus,
			"email":        req.Email,
			"no_telepon":   req.NoTelepon,
			"alamat":       req.Alamat,
			"updated_at":   time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

// ================= SOFT DELETE =================
func (r *alumniRepository) SoftDelete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{"$set": bson.M{
		"is_deleted": true,
		"updated_at": time.Now(),
	}}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

// ================= RESTORE =================
func (r *alumniRepository) Restore(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{"$set": bson.M{
		"is_deleted": false,
		"updated_at": time.Now(),
	}}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

// ================= GET WITHOUT PEKERJAAN =================
func (r *alumniRepository) GetWithoutPekerjaan(ctx context.Context) ([]models.Alumni, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$lookup", Value: bson.M{
			"from":         "pekerjaan_alumni",
			"localField":   "_id",
			"foreignField": "alumni_id",
			"as":           "pekerjaan",
		}}},
		{{Key: "$match", Value: bson.M{
			"pekerjaan":  bson.M{"$size": 0},
			"is_deleted": false,
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var alumniList []models.Alumni
	if err := cursor.All(ctx, &alumniList); err != nil {
		return nil, err
	}
	return alumniList, nil
}

// ================= COUNT WITHOUT PEKERJAAN =================
func (r *alumniRepository) CountWithoutPekerjaan(ctx context.Context) (int, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$lookup", Value: bson.M{
			"from":         "pekerjaan_alumni",
			"localField":   "_id",
			"foreignField": "alumni_id",
			"as":           "pekerjaan",
		}}},
		{{Key: "$match", Value: bson.M{
			"pekerjaan":  bson.M{"$size": 0},
			"is_deleted": false,
		}}},
		{{Key: "$count", Value: "total"}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return 0, err
	}

	if len(results) == 0 {
		return 0, nil
	}

	switch v := results[0]["total"].(type) {
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, nil
	}
}

// ================= SEARCH + SORT + PAGINATION =================
func (r *alumniRepository) GetAlumniRepo(ctx context.Context, search, sortBy, order string, limit, offset int64) ([]models.Alumni, error) {
	filter := bson.M{
		"is_deleted": false,
		"$or": []bson.M{
			{"nim": bson.M{"$regex": search, "$options": "i"}},
			{"nama": bson.M{"$regex": search, "$options": "i"}},
			{"jurusan": bson.M{"$regex": search, "$options": "i"}},
			{"email": bson.M{"$regex": search, "$options": "i"}},
			{"no_telepon": bson.M{"$regex": search, "$options": "i"}},
			{"alamat": bson.M{"$regex": search, "$options": "i"}},
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

	var alumni []models.Alumni
	if err := cursor.All(ctx, &alumni); err != nil {
		return nil, err
	}
	return alumni, nil
}

// ================= COUNT ALUMNI =================
func (r *alumniRepository) CountAlumniRepo(ctx context.Context, search string) (int64, error) {
	filter := bson.M{
		"is_deleted": false,
		"$or": []bson.M{
			{"nim": bson.M{"$regex": search, "$options": "i"}},
			{"nama": bson.M{"$regex": search, "$options": "i"}},
			{"jurusan": bson.M{"$regex": search, "$options": "i"}},
			{"email": bson.M{"$regex": search, "$options": "i"}},
			{"no_telepon": bson.M{"$regex": search, "$options": "i"}},
			{"alamat": bson.M{"$regex": search, "$options": "i"}},
		},
	}
	count, err := r.collection.CountDocuments(ctx, filter)
	return count, err
}
