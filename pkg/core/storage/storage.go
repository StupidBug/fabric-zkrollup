package storage

import (
	"sync"

	"github.com/StupidBug/fabric-zkrollup/pkg/types/evidence"
	"github.com/StupidBug/fabric-zkrollup/pkg/types/status"
)

type Storage struct {
	StateRoot []string
	Evidence  []*evidence.Evidence
	rwmutex   sync.RWMutex
	Cache     map[string]*evidence.Evidence
}

func NewStorage(initialStateRoot string) *Storage {
	return &Storage{
		StateRoot: []string{initialStateRoot},
		Cache:     make(map[string]*evidence.Evidence),
	}
}

func (s *Storage) LastStateRoot() string {
	s.rwmutex.RLock()
	defer s.rwmutex.RUnlock()
	return s.StateRoot[len(s.StateRoot)-1]
}

func (s *Storage) AddBatch(evidences []*evidence.Evidence, stateRoot string) {
	s.rwmutex.Lock()
	defer s.rwmutex.Unlock()
	for _, evidence := range evidences {
		evidence.Status = status.StatusConfirmed
		s.Evidence = append(s.Evidence, evidence)
		s.StateRoot = append(s.StateRoot, stateRoot)
		s.Cache[evidence.Hash] = evidence
	}
}

func (s *Storage) GetEvidenceStatus(hash string) string {
	s.rwmutex.RLock()
	defer s.rwmutex.RUnlock()
	if evidence, ok := s.Cache[hash]; ok {
		return evidence.Status.String()
	}
	return status.StatusPending.String()
}
