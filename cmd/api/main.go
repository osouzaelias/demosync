package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/osouzaelias/demosync/internal/api"
	"github.com/osouzaelias/demosync/internal/kafka"
	"github.com/osouzaelias/demosync/internal/storage"
)

func main() {
	// Inicializa o armazenamento de correlação
	correlationStore := storage.NewCorrelationStore(5 * time.Minute)

	// Detecta e configura o ambiente
	kafkaServer := getKafkaBootstrapServers()
	responseTopic := getKafkaResponseTopic()

	log.Printf("Conectando ao Kafka: %s", kafkaServer)
	log.Printf("Usando tópico de resposta: %s", responseTopic)

	// Inicializa produtor Kafka
	kafkaProducer, err := kafka.NewProducer(kafkaServer)
	if err != nil {
		log.Fatalf("Erro ao criar produtor Kafka: %v", err)
	}
	defer kafkaProducer.Close()

	// Inicializa consumidor Kafka
	kafkaConsumer, err := kafka.NewConsumer(
		kafkaServer,
		responseTopic,
		"payment-response-group",
		correlationStore,
	)
	if err != nil {
		log.Fatalf("Erro ao criar consumidor Kafka: %v", err)
	}

	// Inicia o consumidor em uma goroutine
	ctx, cancel := context.WithCancel(context.Background())
	go kafkaConsumer.Consume(ctx)

	// Inicializa e inicia o servidor HTTP
	server := api.NewServer(kafkaProducer, correlationStore)
	go func() {
		if err := server.Run(":8080"); err != nil {
			log.Fatalf("Erro ao iniciar o servidor: %v", err)
		}
	}()

	log.Println("Servidor iniciado na porta 8080")

	// Espera por sinais de encerramento
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Encerrando aplicação...")
	cancel() // Encerra o consumidor Kafka

	// Espera que o consumidor seja encerrado corretamente
	time.Sleep(1 * time.Second)
}

// getKafkaBootstrapServers obtém os servidores Kafka, adaptando-se ao ambiente
func getKafkaBootstrapServers() string {
	// Verifica a variável de ambiente
	kafkaServer := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")

	// Se não estiver definida ou for o servidor interno do Docker e estivermos rodando localmente
	if kafkaServer == "" {
		return "localhost:29092" // Use a porta externa para o localhost
	}

	// Se estiver definida como kafka:9092 mas estivermos rodando fora do Docker
	if kafkaServer == "kafka:9092" && isRunningLocally() {
		return "localhost:29092" // Use a porta externa para o localhost
	}

	return kafkaServer
}

// getKafkaResponseTopic obtém o tópico de resposta do Kafka
func getKafkaResponseTopic() string {
	responseTopic := os.Getenv("KAFKA_RESPONSE_TOPIC")
	if responseTopic == "" {
		return "payment-responses"
	}
	return responseTopic
}

// isRunningLocally verifica se a aplicação está rodando localmente (fora do Docker)
func isRunningLocally() bool {
	// Verifica se existe o arquivo /.dockerenv que é criado automaticamente em contêineres Docker
	_, err := os.Stat("/.dockerenv")
	return os.IsNotExist(err)
}
