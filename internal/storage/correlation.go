package storage

import (
	"sync"
	"time"

	"github.com/osouzaelias/demosync/pkg/models"
)

// CorrelationStore armazena canais de resposta indexados por ID de correlação
type CorrelationStore struct {
	mu             sync.RWMutex
	channels       map[string]chan *models.CaptureResponse
	expiryDuration time.Duration
	expiries       map[string]time.Time
}

// NewCorrelationStore cria um novo armazenamento de correlação com limpeza automática
func NewCorrelationStore(expiryDuration time.Duration) *CorrelationStore {
	store := &CorrelationStore{
		channels:       make(map[string]chan *models.CaptureResponse),
		expiryDuration: expiryDuration,
		expiries:       make(map[string]time.Time),
	}

	// Inicia a goroutine de limpeza
	go store.cleanupExpired()

	return store
}

// Set armazena um canal de resposta para um ID de correlação
func (s *CorrelationStore) Set(correlationID string, ch chan *models.CaptureResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.channels[correlationID] = ch
	s.expiries[correlationID] = time.Now().Add(s.expiryDuration)
}

// Get obtém um canal de resposta para um ID de correlação
func (s *CorrelationStore) Get(correlationID string) (chan *models.CaptureResponse, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ch, found := s.channels[correlationID]
	return ch, found
}

// Delete remove um canal de resposta do armazenamento
func (s *CorrelationStore) Delete(correlationID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.channels, correlationID)
	delete(s.expiries, correlationID)
}

// cleanupExpired limpa periodicamente os canais expirados
func (s *CorrelationStore) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()

		now := time.Now()
		for id, expiry := range s.expiries {
			if now.After(expiry) {
				// Fecha o canal expirado
				close(s.channels[id])

				// Remove do armazenamento
				delete(s.channels, id)
				delete(s.expiries, id)
			}
		}

		s.mu.Unlock()
	}
}
