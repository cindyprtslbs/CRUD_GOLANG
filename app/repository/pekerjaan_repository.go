package repository

import (
	models "crud-app/app/model"
	"database/sql"
	"fmt"
	"log"
	"time"
)

type PekerjaanRepository interface {
	GetAll() ([]models.Pekerjaan, error)
	GetByID(id int) (*models.Pekerjaan, error)
	GetByAlumniID(alumniID int) ([]models.Pekerjaan, error)
	Create(req models.CreatePekerjaanRequest) (*models.Pekerjaan, error)
	Update(id int, req models.UpdatePekerjaanRequest) (*models.Pekerjaan, error)
	// Delete(id int) error
	SoftDeleteByID(id int) error
	SoftDeleteByOwner(id, alumniID int) error
	RestoreByID(id int) error
	RestoreByOwner(id, alumniID int) error
	GetAlumniIDByPekerjaan(id int) (int, error)
	GetUserByID(id int) (*models.User, error)

	GetBekerjalebih1Tahun() ([]models.AlumniWithPekerjaan, error)
	GetPekerjaanRepo(search, sortBy, order string, limit, offset int) ([]models.Pekerjaan, error)
	CountPekerjaanRepo(search string) (int, error)

	GetTrash() ([]models.Pekerjaan, error)
	GetTrashByOwner(alumniID int) ([]models.Pekerjaan, error)
	Delete(id int, alumniID *int) error
	Restore(id int, alumniID *int) error
}

type pekerjaanRepository struct {
	db *sql.DB
}

func NewPekerjaanRepository(db *sql.DB) PekerjaanRepository {
	return &pekerjaanRepository{db: db}
}

// GET ALL
func (r *pekerjaanRepository) GetAll() ([]models.Pekerjaan, error) {
	rows, err := r.db.Query(`
        SELECT id, alumni_id, nama_perusahaan, posisi_jabatan, bidang_industri, lokasi_kerja, gaji_range, 
		tanggal_mulai_kerja, tanggal_selesai_kerja, status_pekerjaan, deskripsi_pekerjaan, created_at, updated_at
        FROM pekerjaan_alumni 
        WHERE is_deleted = FALSE
        ORDER BY created_at DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Pekerjaan
	for rows.Next() {
		var p models.Pekerjaan
		err := rows.Scan(
			&p.ID, &p.AlumniID, &p.NamaPerusahaan, &p.PosisiJabatan, &p.BidangIndustri,
			&p.LokasiKerja, &p.GajiRange, &p.TanggalMulaiKerja, &p.TanggalSelesaiKerja,
			&p.StatusPekerjaan, &p.DeskripsiPekerjaan, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

// GET BY ID
func (r *pekerjaanRepository) GetByID(id int) (*models.Pekerjaan, error) {
	var p models.Pekerjaan
	row := r.db.QueryRow(`
		SELECT id, alumni_id, nama_perusahaan, posisi_jabatan, bidang_industri, lokasi_kerja,
		       gaji_range, tanggal_mulai_kerja, tanggal_selesai_kerja, status_pekerjaan, 
		       deskripsi_pekerjaan, created_at, updated_at, is_deleted
		FROM pekerjaan_alumni 
		WHERE id=$1
	`, id)

	err := row.Scan(
		&p.ID, &p.AlumniID, &p.NamaPerusahaan, &p.PosisiJabatan, &p.BidangIndustri,
		&p.LokasiKerja, &p.GajiRange, &p.TanggalMulaiKerja, &p.TanggalSelesaiKerja,
		&p.StatusPekerjaan, &p.DeskripsiPekerjaan, &p.CreatedAt, &p.UpdatedAt, &p.IsDeleted,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GET BY ALUMNI ID
func (r *pekerjaanRepository) GetByAlumniID(alumniID int) ([]models.Pekerjaan, error) {
	rows, err := r.db.Query(`
		SELECT id, alumni_id, nama_perusahaan, posisi_jabatan, bidang_industri, lokasi_kerja,
		gaji_range, tanggal_mulai_kerja, tanggal_selesai_kerja, status_pekerjaan, deskripsi_pekerjaan,
		created_at, updated_at
		FROM pekerjaan_alumni 
		WHERE alumni_id=$1 AND is_deleted=FALSE
		ORDER BY created_at DESC`, alumniID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Pekerjaan
	for rows.Next() {
		var p models.Pekerjaan
		if err := rows.Scan(&p.ID, &p.AlumniID, &p.NamaPerusahaan, &p.PosisiJabatan, &p.BidangIndustri,
			&p.LokasiKerja, &p.GajiRange, &p.TanggalMulaiKerja, &p.TanggalSelesaiKerja,
			&p.StatusPekerjaan, &p.DeskripsiPekerjaan, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

// CREATE
func (r *pekerjaanRepository) Create(req models.CreatePekerjaanRequest) (*models.Pekerjaan, error) {
	var id int
	err := r.db.QueryRow(`
		INSERT INTO pekerjaan_alumni (alumni_id, nama_perusahaan, posisi_jabatan, bidang_industri, lokasi_kerja,
		gaji_range, tanggal_mulai_kerja, tanggal_selesai_kerja, status_pekerjaan, deskripsi_pekerjaan, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING id
	`, req.AlumniID, req.NamaPerusahaan, req.PosisiJabatan, req.BidangIndustri, req.LokasiKerja,
		req.GajiRange, req.TanggalMulaiKerja, req.TanggalSelesaiKerja, req.StatusPekerjaan,
		req.DeskripsiPekerjaan, time.Now(), time.Now()).Scan(&id)

	if err != nil {
		return nil, err
	}

	var newPekerjaan models.Pekerjaan
	row := r.db.QueryRow(`
		SELECT id, alumni_id, nama_perusahaan, posisi_jabatan, bidang_industri, lokasi_kerja,
		gaji_range, tanggal_mulai_kerja, tanggal_selesai_kerja, status_pekerjaan, deskripsi_pekerjaan, created_at, updated_at
		FROM pekerjaan_alumni WHERE id=$1
	`, id)

	err = row.Scan(
		&newPekerjaan.ID, &newPekerjaan.AlumniID, &newPekerjaan.NamaPerusahaan, &newPekerjaan.PosisiJabatan, &newPekerjaan.BidangIndustri,
		&newPekerjaan.LokasiKerja, &newPekerjaan.GajiRange, &newPekerjaan.TanggalMulaiKerja, &newPekerjaan.TanggalSelesaiKerja,
		&newPekerjaan.StatusPekerjaan, &newPekerjaan.DeskripsiPekerjaan, &newPekerjaan.CreatedAt, &newPekerjaan.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &newPekerjaan, nil
}

// UPDATE BY ID (ADMIN)
func (r *pekerjaanRepository) Update(id int, req models.UpdatePekerjaanRequest) (*models.Pekerjaan, error) {
	result, err := r.db.Exec(`
		UPDATE pekerjaan_alumni SET nama_perusahaan=$1, posisi_jabatan=$2, bidang_industri=$3, lokasi_kerja=$4,
		gaji_range=$5, tanggal_mulai_kerja=$6, tanggal_selesai_kerja=$7, status_pekerjaan=$8, deskripsi_pekerjaan=$9, updated_at=$10 
		WHERE id=$11
	`, req.NamaPerusahaan, req.PosisiJabatan, req.BidangIndustri, req.LokasiKerja, req.GajiRange,
		req.TanggalMulaiKerja, req.TanggalSelesaiKerja, req.StatusPekerjaan, req.DeskripsiPekerjaan, time.Now(), id)

	if err != nil {
		return nil, err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, sql.ErrNoRows
	}

	var updated models.Pekerjaan
	row := r.db.QueryRow(`
		SELECT id, alumni_id, nama_perusahaan, posisi_jabatan, bidang_industri, lokasi_kerja,
		gaji_range, tanggal_mulai_kerja, tanggal_selesai_kerja, status_pekerjaan, deskripsi_pekerjaan, created_at, updated_at
		FROM pekerjaan_alumni 
		WHERE id=$1
	`, id)

	err = row.Scan(
		&updated.ID, &updated.AlumniID, &updated.NamaPerusahaan, &updated.PosisiJabatan, &updated.BidangIndustri,
		&updated.LokasiKerja, &updated.GajiRange, &updated.TanggalMulaiKerja, &updated.TanggalSelesaiKerja,
		&updated.StatusPekerjaan, &updated.DeskripsiPekerjaan, &updated.CreatedAt, &updated.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

// SOFT DELETE BY ID (ADMIN)
func (r *pekerjaanRepository) SoftDeleteByID(id int) error {
	result, err := r.db.Exec(`
        UPDATE pekerjaan_alumni 
        SET is_deleted = TRUE, updated_at = NOW() 
        WHERE id=$1 AND is_deleted=FALSE
    `, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// SOFT DELETE BY OWNER
func (r *pekerjaanRepository) SoftDeleteByOwner(id, alumniID int) error {
	result, err := r.db.Exec(`
        UPDATE pekerjaan_alumni 
        SET is_deleted = TRUE, updated_at = NOW() 
        WHERE id=$1 AND alumni_id=$2 AND is_deleted=FALSE
    `, id, alumniID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *pekerjaanRepository) GetAlumniIDByPekerjaan(id int) (int, error) {
	var alumniID int
	err := r.db.QueryRow("SELECT alumni_id FROM pekerjaan_alumni WHERE id=$1", id).Scan(&alumniID)
	return alumniID, err
}

// GET USER BY ID
func (r *pekerjaanRepository) GetUserByID(id int) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(`
        SELECT id, username, email, role, alumni_id, created_at
        FROM users
        WHERE id = $1
    `, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Role,
		&user.AlumniID,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// RESTORE
func (r *pekerjaanRepository) RestoreByID(id int) error {
	result, err := r.db.Exec(`
        UPDATE pekerjaan_alumni 
        SET is_deleted = FALSE, updated_at = NOW()
        WHERE id = $1 AND is_deleted = TRUE
    `, id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *pekerjaanRepository) RestoreByOwner(id, alumniID int) error {
	result, err := r.db.Exec(`
        UPDATE pekerjaan_alumni 
        SET is_deleted = FALSE, updated_at = NOW()
        WHERE id = $1 AND alumni_id = $2 AND is_deleted = TRUE
    `, id, alumniID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// BEKERJA LEBIH 1 TAHUN
func (r *pekerjaanRepository) GetBekerjalebih1Tahun() ([]models.AlumniWithPekerjaan, error) {
	rows, err := r.db.Query(`
        SELECT a.id, a.nama, a.jurusan, a.angkatan,
               p.nama_perusahaan, p.posisi_jabatan, p.bidang_industri,
               p.tanggal_mulai_kerja, p.tanggal_selesai_kerja, p.gaji_range
        FROM alumni a
        JOIN pekerjaan_alumni p ON a.id = p.alumni_id
        WHERE (COALESCE(p.tanggal_selesai_kerja, NOW()) - p.tanggal_mulai_kerja) >= INTERVAL '1 year'
          AND p.is_deleted = FALSE
        ORDER BY a.id
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.AlumniWithPekerjaan
	for rows.Next() {
		var row models.AlumniWithPekerjaan
		var jurusan sql.NullString
		var angkatan sql.NullInt64
		var selesai sql.NullTime

		if err := rows.Scan(
			&row.AlumniID, &row.Nama, &jurusan, &angkatan,
			&row.NamaPerusahaan, &row.PosisiJabatan, &row.BidangIndustri,
			&row.TanggalMulaiKerja, &selesai, &row.GajiRange,
		); err != nil {
			return nil, err
		}

		if jurusan.Valid {
			row.Jurusan = jurusan.String
		}
		if angkatan.Valid {
			row.Angkatan = int(angkatan.Int64)
		}
		if selesai.Valid {
			row.TanggalSelesaiKerja = &selesai.Time
		} else {
			row.TanggalSelesaiKerja = nil
		}

		results = append(results, row)
	}

	return results, nil
}

// PAGINATION, SEARCH, SORTING
func (r *pekerjaanRepository) GetPekerjaanRepo(search, sortBy, order string, limit, offset int) ([]models.Pekerjaan, error) {
	query := fmt.Sprintf(`
		SELECT id, alumni_id, nama_perusahaan, posisi_jabatan, bidang_industri, lokasi_kerja, gaji_range, tanggal_mulai_kerja, status_pekerjaan, deskripsi_pekerjaan, created_at, updated_at
		FROM pekerjaan_alumni
		WHERE is_deleted = FALSE AND (
			nama_perusahaan ILIKE $1 
			OR posisi_jabatan ILIKE $1 
			OR bidang_industri ILIKE $1 
			OR lokasi_kerja ILIKE $1
		)
		ORDER BY %s %s
		LIMIT $2 OFFSET $3
	`, sortBy, order)

	rows, err := r.db.Query(query, "%"+search+"%", limit, offset)
	if err != nil {
		log.Println("Query error:", err)
		return nil, err
	}
	defer rows.Close()

	var pekerjaan []models.Pekerjaan
	for rows.Next() {
		var p models.Pekerjaan
		if err := rows.Scan(
			&p.ID, &p.AlumniID, &p.NamaPerusahaan, &p.PosisiJabatan,
			&p.BidangIndustri, &p.LokasiKerja, &p.GajiRange,
			&p.TanggalMulaiKerja, &p.StatusPekerjaan, &p.DeskripsiPekerjaan, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		pekerjaan = append(pekerjaan, p)
	}
	return pekerjaan, nil
}

// COUNT FOR PAGINATION
func (r *pekerjaanRepository) CountPekerjaanRepo(search string) (int, error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM pekerjaan_alumni 
		WHERE is_deleted = FALSE AND (
			nama_perusahaan ILIKE $1 
			OR posisi_jabatan ILIKE $1 
			OR bidang_industri ILIKE $1 
			OR lokasi_kerja ILIKE $1
		)`
	err := r.db.QueryRow(countQuery, "%"+search+"%").Scan(&total)

	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return total, nil
}

// GET TRASH admin bisa lihat semua data di trash, alumni hanya bisa lihat miliknya
func (r *pekerjaanRepository) GetTrash() ([]models.Pekerjaan, error) {
	rows, err := r.db.Query(`
		SELECT id, alumni_id, nama_perusahaan, posisi_jabatan, bidang_industri, lokasi_kerja, gaji_range, 
		       tanggal_mulai_kerja, tanggal_selesai_kerja, status_pekerjaan, deskripsi_pekerjaan, is_deleted, created_at, updated_at
		FROM pekerjaan_alumni 
		WHERE is_deleted = TRUE
		ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Pekerjaan
	for rows.Next() {
		var p models.Pekerjaan
		err := rows.Scan(
			&p.ID, &p.AlumniID, &p.NamaPerusahaan, &p.PosisiJabatan, &p.BidangIndustri,
			&p.LokasiKerja, &p.GajiRange, &p.TanggalMulaiKerja, &p.TanggalSelesaiKerja,
			&p.StatusPekerjaan, &p.DeskripsiPekerjaan, &p.IsDeleted, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func (r *pekerjaanRepository) GetTrashByOwner(alumniID int) ([]models.Pekerjaan, error) {
	rows, err := r.db.Query(`
		SELECT id, alumni_id, nama_perusahaan, posisi_jabatan, bidang_industri, lokasi_kerja, gaji_range,
		       tanggal_mulai_kerja, tanggal_selesai_kerja, status_pekerjaan, deskripsi_pekerjaan, is_deleted, created_at, updated_at
		FROM pekerjaan_alumni 
		WHERE is_deleted = TRUE AND alumni_id = $1
		ORDER BY updated_at DESC
	`, alumniID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Pekerjaan
	for rows.Next() {
		var p models.Pekerjaan
		err := rows.Scan(
			&p.ID, &p.AlumniID, &p.NamaPerusahaan, &p.PosisiJabatan, &p.BidangIndustri,
			&p.LokasiKerja, &p.GajiRange, &p.TanggalMulaiKerja, &p.TanggalSelesaiKerja,
			&p.StatusPekerjaan, &p.DeskripsiPekerjaan, &p.IsDeleted, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

// HARD DELETE DATA
// Admin bisa hard delete semua data, user hanya bisa hard delete datanya sendiri
func (r *pekerjaanRepository) Delete(id int, alumniID *int) error {
	query := `
		DELETE FROM pekerjaan_alumni
		WHERE id = $1
	`

	var result sql.Result
	var err error

	// Jika alumniID tidak nil → user biasa, hanya boleh hapus datanya sendiri
	if alumniID != nil {
		query += " AND alumni_id = $2"
		result, err = r.db.Exec(query, id, *alumniID)
	} else {
		// Jika nil → berarti admin, bisa hapus data siapapun
		result, err = r.db.Exec(query, id)
	}

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}


// RESTORE DATA YANG DI SOFT DELETE
// Admin bisa restore semua data, user hanya bisa restore data miliknya
func (r *pekerjaanRepository) Restore(id int, alumniID *int) error {
	query := `
		UPDATE pekerjaan_alumni
		SET is_deleted = FALSE, updated_at = NOW()
		WHERE id = $1
	`

	// Jika alumniID tidak nil → berarti user biasa (restore hanya miliknya)
	// Jika alumniID nil → berarti admin (restore semua bisa)
	var result sql.Result
	var err error

	if alumniID != nil {
		query += " AND alumni_id = $2"
		result, err = r.db.Exec(query, id, *alumniID)
	} else {
		result, err = r.db.Exec(query, id)
	}

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

			
