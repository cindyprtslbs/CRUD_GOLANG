package repository

import (
	models "crud-app/app/model"
	"database/sql"
	"fmt"
	"log"
	"time"
	// "golang.org/x/text/search"
)

type AlumniRepository interface {
	GetAll() ([]models.Alumni, error)
	GetByID(id int) (*models.Alumni, error)
	Create(req models.CreateAlumniRequest) (*models.Alumni, error)
	Update(id int, req models.UpdateAlumniRequest) (*models.Alumni, error)
	// Delete(id int) error
	SoftDelete(id, userID int, role string) error
	Restore(alumniID int, userID int) error 

	CountWithoutPekerjaan() (int, error)
	GetWithoutPekerjaan() ([]models.Alumni, error)

	GetAlumniRepo(search, sortBy, order string, limit, offset int) ([]models.Alumni, error)
	CountAlumniRepo(search string) (int, error)
}

type alumniRepository struct {
	db *sql.DB
}

func NewAlumniRepository(db *sql.DB) AlumniRepository {
	return &alumniRepository{db: db}
}

func (r *alumniRepository) GetAll() ([]models.Alumni, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, nim, nama, jurusan, angkatan, tahun_lulus, email, no_telepon, alamat, created_at, updated_at
		FROM alumni
		WHERE is_deleted = FALSE
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Alumni
	for rows.Next() {
		var a models.Alumni
		if err := rows.Scan(&a.ID, &a.UserID, &a.NIM, &a.Nama, &a.Jurusan, &a.Angkatan, &a.TahunLulus,
			&a.Email, &a.NoTelepon, &a.Alamat, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

func (r *alumniRepository) GetByID(id int) (*models.Alumni, error) {
	var a models.Alumni
	row := r.db.QueryRow(`
		SELECT id, user_id, nim, nama, jurusan, angkatan, tahun_lulus,
			email, no_telepon, alamat, created_at, updated_at, is_deleted
		FROM alumni WHERE id=$1`, id)

	err := row.Scan(
		&a.ID,
		&a.UserID,
		&a.NIM,
		&a.Nama,
		&a.Jurusan,
		&a.Angkatan,
		&a.TahunLulus,
		&a.Email,
		&a.NoTelepon,
		&a.Alamat,
		&a.CreatedAt,
		&a.UpdatedAt,
		&a.IsDeleted,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *alumniRepository) Create(req models.CreateAlumniRequest) (*models.Alumni, error) {
	var id int
	err := r.db.QueryRow(`
		INSERT INTO alumni (user_id, nim, nama, jurusan, angkatan, tahun_lulus, email, no_telepon, alamat, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id
	`, req.UserID, req.NIM, req.Nama, req.Jurusan, req.Angkatan, req.TahunLulus,
		req.Email, req.NoTelepon, req.Alamat, time.Now(), time.Now()).Scan(&id)
	if err != nil {
		return nil, err
	}
	var newAlumni models.Alumni
	row := r.db.QueryRow(`
		SELECT id, user_id, nim, nama, jurusan, angkatan, tahun_lulus, email, no_telepon, alamat, created_at, updated_at
		FROM alumni 
		WHERE id=$1
	`, id)

	err = row.Scan(
		&newAlumni.ID, &newAlumni.UserID, &newAlumni.NIM, &newAlumni.Nama, &newAlumni.Jurusan,
		&newAlumni.Angkatan, &newAlumni.TahunLulus, &newAlumni.Email, &newAlumni.NoTelepon,
		&newAlumni.Alamat, &newAlumni.CreatedAt, &newAlumni.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &newAlumni, nil
}

func (r *alumniRepository) Update(id int, req models.UpdateAlumniRequest) (*models.Alumni, error) {
	result, err := r.db.Exec(`
		UPDATE alumni 
		SET user_id=$1, nim=$2, nama=$3, jurusan=$4, angkatan=$5, tahun_lulus=$6, email=$7, no_telepon=$8, alamat=$9, updated_at=$10
		WHERE id=$11
	`, req.UserID, req.NIM, req.Nama, req.Jurusan, req.Angkatan, req.TahunLulus, req.Email, req.NoTelepon, req.Alamat, time.Now(), id)
	if err != nil {
		return nil, err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, sql.ErrNoRows
	}
	var updated models.Alumni
	row := r.db.QueryRow(`
		SELECT id, user_id, nim, nama, jurusan, angkatan, tahun_lulus, email, no_telepon, alamat, created_at, updated_at
		FROM alumni 
		WHERE id = $1
	`, id)

	err = row.Scan(
		&updated.ID, &updated.UserID, &updated.NIM, &updated.Nama, &updated.Jurusan, &updated.Angkatan, &updated.TahunLulus,
		&updated.Email, &updated.NoTelepon, &updated.Alamat, &updated.CreatedAt, &updated.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}
	return &updated, nil
}

// func (r *alumniRepository) Delete(id int) error {
// 	result, err := r.db.Exec(`DELETE FROM alumni WHERE id=$1`, id)
// 	if err != nil {
// 		return err
// 	}
// 	rowsAffected, _ := result.RowsAffected()
// 	if rowsAffected == 0 {
// 		return sql.ErrNoRows
// 	}
// 	return nil
// }

// =================== SOFT DELETE ===================
func (r *alumniRepository) SoftDelete(alumniID int, userID int, role string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	var targetUserID int
	err = tx.QueryRow(`SELECT user_id FROM alumni WHERE id = $1`, alumniID).Scan(&targetUserID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("alumni tidak ditemukan: %v", err)
	}

	// Admin bisa hapus siapa pun, alumni hanya bisa hapus miliknya sendiri
	if role == "admin" {
		_, err = tx.Exec(`
			UPDATE alumni 
			SET is_deleted = TRUE, updated_at = NOW()
			WHERE id = $1 AND is_deleted = FALSE
		`, alumniID)
	} else {
		_, err = tx.Exec(`
			UPDATE alumni 
			SET is_deleted = TRUE, updated_at = NOW()
			WHERE id = $1 AND user_id = $2 AND is_deleted = FALSE
		`, alumniID, userID)
	}
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal menghapus alumni: %v", err)
	}

	// Nonaktifkan user terkait alumni
	_, err = tx.Exec(`
		UPDATE users 
		SET is_active = FALSE, updated_at = NOW()
		WHERE id = $1
	`, targetUserID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal menonaktifkan user alumni: %v", err)
	}

	return tx.Commit()
}


// =================== RESTORE ===================
func (r *alumniRepository) Restore(alumniID int, userID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	// Restore alumni
	_, err = tx.Exec(`
		UPDATE alumni 
		SET is_deleted = FALSE 
		WHERE id = $1
	`, alumniID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal merestore alumni: %w", err)
	}

	// Aktifkan kembali user terkait
	_, err = tx.Exec(`
		UPDATE users 
		SET is_active = TRUE 
		WHERE id = $1
	`, userID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal mengaktifkan user alumni: %w", err)
	}

	// Commit perubahan
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}


// =================== WITHOUT PEKERJAAN ===================
func (r *alumniRepository) CountWithoutPekerjaan() (int, error) {
	var count int
	err := r.db.QueryRow(`
        SELECT COUNT(*) AS jumlah_alumni_tanpa_pekerjaan
		FROM alumni a
		LEFT JOIN pekerjaan_alumni p
			ON a.id = p.alumni_id
		WHERE p.alumni_id IS NULL;
    `).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *alumniRepository) GetWithoutPekerjaan() ([]models.Alumni, error) {
	rows, err := r.db.Query(`
		SELECT a.id, a.user_id, a.nim, a.nama, a.jurusan, a.angkatan, 
			a.tahun_lulus, a.email, a.no_telepon, a.alamat, 
			a.created_at, a.updated_at
		FROM alumni a
		LEFT JOIN pekerjaan_alumni p ON a.id = p.alumni_id
		WHERE p.alumni_id IS NULL
	`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Alumni
	for rows.Next() {
		var a models.Alumni
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.NIM, &a.Nama, &a.Jurusan, &a.Angkatan,
			&a.TahunLulus, &a.Email, &a.NoTelepon, &a.Alamat,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}

		list = append(list, a)
	}
	return list, nil
}

func (r *alumniRepository) GetAlumniRepo(search, sortBy, order string, limit, offset int) ([]models.Alumni, error) {
	query := fmt.Sprintf(`
		SELECT id, user_id, nim, nama, jurusan, angkatan, tahun_lulus, 
			email, no_telepon, alamat, created_at, updated_at
		FROM alumni
		WHERE nama ILIKE $1 
		OR nim ILIKE $1 
		OR email ILIKE $1 
		OR jurusan ILIKE $1 
		OR CAST(angkatan AS TEXT) ILIKE $1 
		OR CAST(tahun_lulus AS TEXT) ILIKE $1
		OR no_telepon ILIKE $1
		OR alamat ILIKE $1
		ORDER BY %s %s
		LIMIT $2 OFFSET $3
	`, sortBy, order)

	rows, err := r.db.Query(query, "%"+search+"%", limit, offset)
	if err != nil {
		log.Println("Query error:", err)
		return nil, err
	}
	defer rows.Close()

	var alumni []models.Alumni
	for rows.Next() {
		var a models.Alumni
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.NIM, &a.Nama, &a.Jurusan,
			&a.Angkatan, &a.TahunLulus, &a.Email, &a.NoTelepon,
			&a.Alamat, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}

		alumni = append(alumni, a)
	}
	return alumni, nil
}

func (r *alumniRepository) CountAlumniRepo(search string) (int, error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM alumni
	WHERE nama ILIKE $1 
		OR nim ILIKE $1 
		OR email ILIKE $1 
		OR jurusan ILIKE $1 
		OR CAST(angkatan AS TEXT) ILIKE $1 
		OR CAST(tahun_lulus AS TEXT) ILIKE $1
		OR no_telepon ILIKE $1
		OR alamat ILIKE $1`
	err := r.db.QueryRow(countQuery, "%"+search+"%").Scan(&total)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return total, nil
}
