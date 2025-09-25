package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/osouzaelias/demosync/pkg/models"
)

func main() {
	log.Println("Iniciando simulador de processamento de pagamentos...")

	// Inicializa o gerador de números aleatórios
	rand.Seed(time.Now().UnixNano())

	// Obtém configurações do ambiente
	bootstrapServers := getEnv("KAFKA_BOOTSTRAP_SERVERS", "kafka:9092")
	requestTopic := getEnv("KAFKA_REQUEST_TOPIC", "payment-requests")
	responseTopic := getEnv("KAFKA_RESPONSE_TOPIC", "payment-responses")
	processingDelayMs, _ := strconv.Atoi(getEnv("PROCESSING_DELAY_MS", "2000"))
	processingDelay := time.Duration(processingDelayMs) * time.Millisecond

	log.Printf("Configuração: Kafka=%s, Tópico de Requisição=%s, Tópico de Resposta=%s, Atraso=%v",
		bootstrapServers, requestTopic, responseTopic, processingDelay)

	// Configura o consumidor Kafka
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  bootstrapServers,
		"group.id":           "payment-processor-group",
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": true,
	})
	if err != nil {
		log.Fatalf("Erro ao criar consumidor Kafka: %v", err)
	}
	defer consumer.Close()

	// Configura o produtor Kafka
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"client.id":         "payment-processor",
	})
	if err != nil {
		log.Fatalf("Erro ao criar produtor Kafka: %v", err)
	}
	defer producer.Close()

	// Inscreve-se no tópico de requisições
	if err := consumer.SubscribeTopics([]string{requestTopic}, nil); err != nil {
		log.Fatalf("Erro ao se inscrever no tópico %s: %v", requestTopic, err)
	}
	log.Printf("Inscrito no tópico de requisições: %s", requestTopic)

	// Canal para tratar sinais de encerramento
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Contexto para controlar o ciclo de vida
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Inicia o processamento em uma goroutine
	go processMessages(ctx, consumer, producer, responseTopic, processingDelay)

	// Aguarda sinal de encerramento
	<-sigChan
	log.Println("Encerrando o simulador...")
}

func processMessages(ctx context.Context, consumer *kafka.Consumer, producer *kafka.Producer, responseTopic string, processingDelay time.Duration) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := consumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				// Ignora timeouts
				if err.(kafka.Error).Code() != kafka.ErrTimedOut {
					log.Printf("Erro ao ler mensagem: %v", err)
				}
				continue
			}

			// Processa a mensagem
			log.Printf("Mensagem recebida: %s", string(msg.Value))

			// Desserializa a requisição
			var request models.CaptureRequest
			if err := json.Unmarshal(msg.Value, &request); err != nil {
				log.Printf("Erro ao deserializar requisição: %v", err)
				continue
			}

			// Simula tempo de processamento
			log.Printf("Processando requisição para transação %s...", request.TransactionID)
			time.Sleep(processingDelay)

			// Gera resposta (sucesso ou falha)
			response := generateResponse(request)
			log.Printf("Resposta gerada: %+v", response)

			// Serializa a resposta
			responseJSON, err := json.Marshal(response)
			if err != nil {
				log.Printf("Erro ao serializar resposta: %v", err)
				continue
			}

			// Envia a resposta
			err = producer.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic:     &responseTopic,
					Partition: kafka.PartitionAny,
				},
				Key:   msg.Key, // Usa o mesmo ID de correlação
				Value: responseJSON,
			}, nil)

			if err != nil {
				log.Printf("Erro ao enviar resposta: %v", err)
			} else {
				log.Printf("Resposta enviada para o tópico %s", responseTopic)
			}
		}
	}
}

func generateResponse(request models.CaptureRequest) models.CaptureResponse {
	// Simula uma taxa de aprovação de 80%
	isApproved := rand.Float32() < 0.8

	response := models.CaptureResponse{
		CorrelationID:  request.CorrelationID,
		TransactionID:  request.TransactionID,
	}

	if isApproved {
		response.Status = "approved"
		response.AuthorizationCode = generateAuthCode()
	} else {
		response.Status = "declined"
		response.ErrorCode = "payment_declined"
		response.ErrorMessage = "Pagamento recusado pela operadora"
	}

	return response
}

func generateAuthCode() string {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	code := make([]byte, 6)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}
	return string(code)
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
