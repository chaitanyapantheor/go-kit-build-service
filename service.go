package gokitbuildservice

import (
	"context"
	"errors"
	"sync"
)

// Build represents a single cloud build.
// ID should be globally unique.
type Build struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// Service is a simple CRUD interface for user profiles.
type Service interface {
	PostBuild(ctx context.Context, b Build) error
	GetBuild(ctx context.Context, id string) (Build, error)
	PutBuild(ctx context.Context, id string, b Build) error
	PatchBuild(ctx context.Context, id string, b Build) error
	DeleteBuild(ctx context.Context, id string) error
}

var (
	ErrInconsistentIDs = errors.New("inconsistent IDs")
	ErrAlreadyExists   = errors.New("already exists")
	ErrNotFound        = errors.New("not found")
)

type inmemService struct {
	mtx sync.RWMutex
	m   map[string]Build
}

func NewInmemService() Service {
	return &inmemService{
		m: map[string]Build{},
	}
}

func (s *inmemService) PostBuild(ctx context.Context, b Build) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if _, ok := s.m[b.ID]; ok {
		return ErrAlreadyExists // POST = create, don't overwrite
	}
	s.m[b.ID] = b
	return nil
}

func (s *inmemService) GetBuild(ctx context.Context, id string) (Build, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	b, ok := s.m[id]
	if !ok {
		return Build{}, ErrNotFound
	}
	return b, nil
}

func (s *inmemService) PutBuild(ctx context.Context, id string, b Build) error {
	if id != b.ID {
		return ErrInconsistentIDs
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.m[id] = b // PUT = create or update
	return nil
}

func (s *inmemService) PatchBuild(ctx context.Context, id string, b Build) error {
	if b.ID != "" && id != b.ID {
		return ErrInconsistentIDs
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	existing, ok := s.m[id]
	if !ok {
		return ErrNotFound // PATCH = update existing, don't create
	}

	// We assume that it's not possible to PATCH the ID, and that it's not
	// possible to PATCH any field to its zero value. That is, the zero value
	// means not specified. The way around this is to use e.g. Name *string in
	// the Profile definition. But since this is just a demonstrative example,
	// I'm leaving that out.

	if b.Name != "" {
		existing.Name = b.Name
	}
	s.m[id] = existing
	return nil
}

func (s *inmemService) DeleteBuild(ctx context.Context, id string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if _, ok := s.m[id]; !ok {
		return ErrNotFound
	}
	delete(s.m, id)
	return nil
}
