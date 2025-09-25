package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/osouzaelias/demosync/internal/storage"
	"github.com/osouzaelias/demosync/pkg/models"
)

// Consumer encapsula a funcionalidade do consumidor Kafka
type Consumer struct {
	consumer         *kafka.Consumer
	correlationStore *storage.CorrelationStore
}

// NewConsumer cria uma nova instância do consumidor Kafka
func NewConsumer(bootstrapServers, topic, groupID string, correlationStore *storage.CorrelationStore) (*Consumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  bootstrapServers,
		"group.id":           groupID,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": true,
	})

	if err != nil {
		return nil, err
	}

	// Inscreve-se no tópico
	err = c.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		c.Close()
		return nil, err
	}

	return &Consumer{
		consumer:         c,
		correlationStore: correlationStore,
	}, nil
}

// Consume inicia o consumo de mensagens do Kafka
func (c *Consumer) Consume(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Encerrando consumidor Kafka...")
				c.consumer.Close()
				return

			default:
				msg, err := c.consumer.ReadMessage(100 * time.Millisecond)
				if err != nil {
					// Timeout ou outro erro temporário
					if err.(kafka.Error).Code() != kafka.ErrTimedOut {
						log.Printf("Erro ao consumir mensagem: %v\n", err)
					}
					continue
				}

				// Processa a mensagem
				c.processMessage(msg)
			}
		}
	}()
}

// processMessage processa uma mensagem recebida do Kafka
func (c *Consumer) processMessage(msg *kafka.Message) {
	// Extrai o ID de correlação da chave da mensagem
	correlationID := string(msg.Key)

	// Busca o canal de resposta do armazenamento
	responseChan, found := c.correlationStore.Get(correlationID)
	if !found {
		log.Printf("Resposta recebida para ID de correlação desconhecido: %s\n", correlationID)
		return
	}

	// Deserializa a resposta
	var response models.CaptureResponse
	if err := json.Unmarshal(msg.Value, &response); err != nil {
		log.Printf("Erro ao deserializar resposta: %v\n", err)
		return
	}

	// Envia a resposta para o canal
	responseChan <- &response

	// Remove o canal do armazenamento
	c.correlationStore.Delete(correlationID)
}
