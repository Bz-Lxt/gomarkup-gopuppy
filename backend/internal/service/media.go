package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"gopuppy/internal/clock"
	"gopuppy/internal/domain"
	"gopuppy/internal/repo"
	"gopuppy/internal/storage"
)

type Media struct {
	Repo    *repo.Media
	Family  *Family
	Store   storage.Store
}

func (s *Media) Upload(ctx context.Context, userID, petID uuid.UUID, kind domain.MediaKind, filename string, r io.Reader) (*domain.MediaFile, error) {
	p, _, err := s.Family.MustWritePet(ctx, userID, petID)
	if err != nil {
		return nil, err
	}
	if !kind.Valid() {
		return nil, fmt.Errorf("%w: kind", domain.ErrValidation)
	}
	filename = filepath.Base(filename)
	if err := storage.SanitizeFilename(filename); err != nil {
		return nil, err
	}
	sha, data, err := storage.HashReader(r)
	if err != nil {
		return nil, err
	}
	if exist, err := s.Repo.BySHA(ctx, p.FamilyID, sha); err == nil {
		return exist, nil
	}
	head := data
	if len(head) > 16 {
		head = data[:16]
	}
	mime, err := storage.Sniff(head)
	if err != nil {
		return nil, err
	}
	if kind == domain.MediaReportPDF && mime != "application/pdf" {
		return nil, fmt.Errorf("%w: report must be pdf", domain.ErrValidation)
	}
	if kind == domain.MediaPhoto && !strings.HasPrefix(mime, "image/") {
		return nil, fmt.Errorf("%w: photo must be image", domain.ErrValidation)
	}
	key := storage.BuildKey(petID.String(), string(kind), sha, storage.ExtForMIME(mime), clock.Now())
	if err := s.Store.Put(ctx, key, bytes.NewReader(data), int64(len(data)), mime); err != nil {
		return nil, err
	}
	m := &domain.MediaFile{
		ID: uuid.New(), FamilyID: p.FamilyID, PetID: petID, Kind: kind,
		StorageDriver: s.Store.Driver(), ObjectKey: key, Filename: filename,
		MIME: mime, SizeBytes: int64(len(data)), SHA256: sha, UploadedBy: userID, CreatedAt: clock.Now(),
	}
	if err := s.Repo.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *Media) List(ctx context.Context, userID, petID uuid.UUID, kind string) ([]domain.MediaFile, error) {
	if _, _, err := s.Family.MustReadPet(ctx, userID, petID); err != nil {
		return nil, err
	}
	if kind != "" && !domain.MediaKind(kind).Valid() {
		return nil, fmt.Errorf("%w: kind", domain.ErrValidation)
	}
	return s.Repo.ListByPet(ctx, petID, kind)
}

func (s *Media) Open(ctx context.Context, userID, mediaID uuid.UUID) (*domain.MediaFile, io.ReadCloser, error) {
	m, err := s.Repo.ByID(ctx, mediaID)
	if err != nil {
		return nil, nil, domain.ErrNotFound
	}
	if _, _, err := s.Family.MustReadPet(ctx, userID, m.PetID); err != nil {
		return nil, nil, domain.ErrNotFound
	}
	rc, _, err := s.Store.Get(ctx, m.ObjectKey)
	if err != nil {
		return nil, nil, err
	}
	return m, rc, nil
}
