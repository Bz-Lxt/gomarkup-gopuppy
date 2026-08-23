package repo

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopuppy/internal/clock"
	"gopuppy/internal/domain"
)

type Pets struct{ Pool *pgxpool.Pool }

func (r *Pets) Create(ctx context.Context, p *domain.Pet) error {
	_, err := r.Pool.Exec(ctx, `INSERT INTO pets(id,family_id,name,species,breed,gender,birthday,avatar_key,neutered,chip_no,weight_min,weight_max,note,created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		p.ID, p.FamilyID, p.Name, p.Species, p.Breed, p.Gender, clock.DateOf(p.Birthday),
		p.AvatarKey, p.Neutered, p.ChipNo, p.WeightMin, p.WeightMax, p.Note, p.CreatedAt)
	return err
}

func (r *Pets) Update(ctx context.Context, p *domain.Pet) error {
	_, err := r.Pool.Exec(ctx, `UPDATE pets SET name=$2,species=$3,breed=$4,gender=$5,birthday=$6,avatar_key=$7,neutered=$8,chip_no=$9,weight_min=$10,weight_max=$11,note=$12
		WHERE id=$1 AND archived_at IS NULL`,
		p.ID, p.Name, p.Species, p.Breed, p.Gender, clock.DateOf(p.Birthday),
		p.AvatarKey, p.Neutered, p.ChipNo, p.WeightMin, p.WeightMax, p.Note)
	return err
}

func (r *Pets) Archive(ctx context.Context, id uuid.UUID, at time.Time) error {
	_, err := r.Pool.Exec(ctx, `UPDATE pets SET archived_at=$2 WHERE id=$1`, id, at)
	return err
}

func (r *Pets) ByID(ctx context.Context, id uuid.UUID) (*domain.Pet, error) {
	p, err := r.scan(ctx, `SELECT id,family_id,name,species,breed,gender,birthday,avatar_key,neutered,chip_no,weight_min,weight_max,note,archived_at,created_at
		FROM pets WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	p.Age = domain.CalcAge(p.Birthday, clock.Now())
	return p, nil
}

func (r *Pets) ListByFamily(ctx context.Context, familyID uuid.UUID, includeArchived bool) ([]domain.Pet, error) {
	q := `SELECT id,family_id,name,species,breed,gender,birthday,avatar_key,neutered,chip_no,weight_min,weight_max,note,archived_at,created_at
		FROM pets WHERE family_id=$1`
	if !includeArchived {
		q += ` AND archived_at IS NULL`
	}
	q += ` ORDER BY created_at`
	rows, err := r.Pool.Query(ctx, q, familyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	now := clock.Now()
	var out []domain.Pet
	for rows.Next() {
		p, err := scanPet(rows)
		if err != nil {
			return nil, err
		}
		p.Age = domain.CalcAge(p.Birthday, now)
		out = append(out, *p)
	}
	return out, rows.Err()
}

type petScanner interface {
	Scan(dest ...any) error
}

func (r *Pets) scan(ctx context.Context, q string, arg any) (*domain.Pet, error) {
	row := r.Pool.QueryRow(ctx, q, arg)
	p, err := scanPet(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return p, err
}

func scanPet(s petScanner) (*domain.Pet, error) {
	var p domain.Pet
	err := s.Scan(&p.ID, &p.FamilyID, &p.Name, &p.Species, &p.Breed, &p.Gender, &p.Birthday,
		&p.AvatarKey, &p.Neutered, &p.ChipNo, &p.WeightMin, &p.WeightMax, &p.Note, &p.ArchivedAt, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
