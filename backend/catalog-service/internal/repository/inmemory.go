package repository

import "strings"

// Track stores catalog metadata used by command pre-check paths.
type Track struct {
	TrackID    string
	Title      string
	Artist     string
	IsPlayable bool
	ReasonCode string
}

// Store exposes track lookup behavior.
type Store interface {
	FindByID(trackID string) (Track, bool)
}

// InMemoryStore stores static track fixtures for local and test use.
type InMemoryStore struct {
	tracks map[string]Track
}

// NewInMemoryStore creates a seeded catalog repository.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		tracks: map[string]Track{
			"trk_1": {
				TrackID:    "trk_1",
				Title:      "Song One",
				Artist:     "Artist One",
				IsPlayable: true,
			},
			"trk_2": {
				TrackID:    "trk_2",
				Title:      "Song Two",
				Artist:     "Artist Two",
				IsPlayable: false,
				ReasonCode: "license_blocked",
			},
		},
	}
}

// FindByID returns one track record by ID.
func (s *InMemoryStore) FindByID(trackID string) (Track, bool) {
	track, ok := s.tracks[strings.TrimSpace(trackID)]
	return track, ok
}
