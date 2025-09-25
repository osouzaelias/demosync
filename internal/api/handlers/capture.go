package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/osouzaelias/demosync/internal/kafka"
	"github.com/osouzaelias/demosync/internal/storage"
	"github.com/osouzaelias/demosync/pkg/models"
)

// CaptureHandler manipula requisições de captura de pagamento
type CaptureHandler struct {
	kafkaProducer    *kafka.Producer
	correlationStore *storage.CorrelationStore
}

// NewCaptureHandler cria um novo handler de captura
func NewCaptureHandler(kafkaProducer *kafka.Producer, correlationStore *storage.CorrelationStore) *CaptureHandler {
	return &CaptureHandler{
		kafkaProducer:    kafkaProducer,
		correlationStore: correlationStore,
	}
}

// Capture processa uma solicitação de captura de pagamento
func (h *CaptureHandler) Capture(c *gin.Context) {
	var request models.CaptureRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Requisição inválida"})
		return
	}

	// Gera um ID de correlação único para rastrear a solicitação
	correlationID := uuid.New().String()

	// Cria um canal para receber a resposta
	responseChan := make(chan *models.CaptureResponse)

	// Registra o canal no armazenamento de correlação
	h.correlationStore.Set(correlationID, responseChan)

	// Adiciona o ID de correlação à solicitação
	request.CorrelationID = correlationID

	// Serializa a solicitação para JSON
	requestJSON, err := json.Marshal(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao serializar solicitação"})
		return
	}

	// Envia a solicitação para o Kafka
	err = h.kafkaProducer.Produce("payment-requests", correlationID, requestJSON)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao enviar solicitação"})
		return
	}

	// Aguarda a resposta com timeout
	select {
	case response := <-responseChan:
		// Resposta recebida
		c.JSON(http.StatusOK, response)

	case <-time.After(30 * time.Second):
		// Timeout - nenhuma resposta recebida em tempo hábil
		h.correlationStore.Delete(correlationID)
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Tempo limite excedido esperando resposta"})
	}
}
