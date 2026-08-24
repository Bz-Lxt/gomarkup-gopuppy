package repo

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopuppy/internal/domain"
)

type Media struct{ Pool *pgxpool.Pool }

func (r *Media) Create(ctx context.Context, m *domain.MediaFile) error {
	_, err := r.Pool.Exec(ctx, `INSERT INTO media_files(id,family_id,pet_id,kind,storage_driver,object_key,filename,mime,size_bytes,sha256,uploaded_by,created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		m.ID, m.FamilyID, m.PetID, m.Kind, m.StorageDriver, m.ObjectKey, m.Filename, m.MIME, m.SizeBytes, m.SHA256, m.UploadedBy, m.CreatedAt)
	return err
}

func (r *Media) BySHA(ctx context.Context, familyID uuid.UUID, sha string) (*domain.MediaFile, error) {
	return r.scan(ctx, `SELECT id,family_id,pet_id,kind,storage_driver,object_key,filename,mime,size_bytes,sha256,uploaded_by,created_at
		FROM media_files WHERE family_id=$1 AND sha256=$2`, familyID, sha)
}

func (r *Media) ByID(ctx context.Context, id uuid.UUID) (*domain.MediaFile, error) {
	m, err := scanMedia(r.Pool.QueryRow(ctx, `SELECT id,family_id,pet_id,kind,storage_driver,object_key,filename,mime,size_bytes,sha256,uploaded_by,created_at
		FROM media_files WHERE id=$1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return m, err
}

func (r *Media) ListByPet(ctx context.Context, petID uuid.UUID, kind string) ([]domain.MediaFile, error) {
	q := `SELECT id,family_id,pet_id,kind,storage_driver,object_key,filename,mime,size_bytes,sha256,uploaded_by,created_at
		FROM media_files WHERE pet_id=$1`
	args := []any{petID}
	if kind != "" {
		q += ` AND kind=$2`
		args = append(args, kind)
	}
	q += ` ORDER BY created_at DESC`
	rows, err := r.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.MediaFile
	for rows.Next() {
		m, err := scanMedia(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *m)
	}
	return out, rows.Err()
}

func (r *Media) InsertMockDelivery(ctx context.Context, channel, payload string, at time.Time) error {
	_, err := r.Pool.Exec(ctx, `INSERT INTO mock_deliveries(id,channel,payload,created_at) VALUES ($1,$2,$3,$4)`,
		uuid.New(), channel, payload, at)
	return err
}

func (r *Media) scan(ctx context.Context, q string, a, b any) (*domain.MediaFile, error) {
	var row pgx.Row
	if b == nil {
		row = r.Pool.QueryRow(ctx, q, a)
	} else {
		row = r.Pool.QueryRow(ctx, q, a, b)
	}
	m, err := scanMedia(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return m, err
}

type mediaScanner interface {
	Scan(dest ...any) error
}

func scanMedia(s mediaScanner) (*domain.MediaFile, error) {
	var m domain.MediaFile
	err := s.Scan(&m.ID, &m.FamilyID, &m.PetID, &m.Kind, &m.StorageDriver, &m.ObjectKey, &m.Filename, &m.MIME, &m.SizeBytes, &m.SHA256, &m.UploadedBy, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}
