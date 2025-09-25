package models

// CaptureRequest representa uma solicitação de captura de pagamento
type CaptureRequest struct {
	// ID de correlação para rastrear a solicitação/resposta
	CorrelationID string `json:"correlation_id,omitempty"`

	// ID da transação
	TransactionID string `json:"transaction_id" binding:"required"`

	// Valor a ser capturado
	Amount float64 `json:"amount" binding:"required"`

	// Moeda (ex: BRL, USD)
	Currency string `json:"currency" binding:"required"`

	// Informações adicionais
	Metadata map[string]string `json:"metadata,omitempty"`
}

// CaptureResponse representa uma resposta de captura de pagamento
type CaptureResponse struct {
	// ID de correlação
	CorrelationID string `json:"correlation_id"`

	// ID da transação
	TransactionID string `json:"transaction_id"`

	// Status da transação
	Status string `json:"status"`

	// Código de autorização
	AuthorizationCode string `json:"authorization_code,omitempty"`

	// Código de erro (se houver)
	ErrorCode string `json:"error_code,omitempty"`

	// Mensagem de erro (se houver)
	ErrorMessage string `json:"error_message,omitempty"`
}
