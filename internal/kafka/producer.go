package kafka

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// Producer encapsula a funcionalidade do produtor Kafka
type Producer struct {
	producer *kafka.Producer
}

// NewProducer cria uma nova instância do produtor Kafka
func NewProducer(bootstrapServers string) (*Producer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"client.id":         "payment-api",
		"acks":              "all",
	})

	if err != nil {
		return nil, err
	}

	return &Producer{producer: p}, nil
}

// Produce envia uma mensagem para um tópico Kafka
func (p *Producer) Produce(topic string, key string, value []byte) error {
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(key),
		Value: value,
	}

	// Entrega a mensagem para o canal de mensagens do produtor
	return p.producer.Produce(message, nil)
}

// Close fecha o produtor Kafka
func (p *Producer) Close() {
	p.producer.Close()
}
