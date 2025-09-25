package api

import (
	"github.com/gin-gonic/gin"
	"github.com/osouzaelias/demosync/internal/api/handlers"
	"github.com/osouzaelias/demosync/internal/kafka"
	"github.com/osouzaelias/demosync/internal/storage"
)

// Server representa o servidor HTTP da API
type Server struct {
	router           *gin.Engine
	kafkaProducer    *kafka.Producer
	correlationStore *storage.CorrelationStore
}

// NewServer cria uma nova instância do servidor
func NewServer(kafkaProducer *kafka.Producer, correlationStore *storage.CorrelationStore) *Server {
	server := &Server{
		router:           gin.Default(),
		kafkaProducer:    kafkaProducer,
		correlationStore: correlationStore,
	}

	// Configura as rotas
	server.setupRoutes()

	return server
}

// setupRoutes configura as rotas da API
func (s *Server) setupRoutes() {
	// Handler para a rota de captura
	captureHandler := handlers.NewCaptureHandler(s.kafkaProducer, s.correlationStore)

	// Grupo de rotas para API
	api := s.router.Group("/api/v1")
	{
		api.POST("/payments/capture", captureHandler.Capture)
	}
}

// Run inicia o servidor HTTP
func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}
