package rabbit_mq

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

const exchangeName = "email_service_topics"

type PublisherManager struct {
	conn                       *amqp.Connection
	ch                         *amqp.Channel
	GenericEmailQueuePublisher *GenericEmailPublisher
}

func NewPublisherManager() PublisherManager {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	log.Printf("Connected to RabbitMQ")

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	log.Printf("Opened a channel")

	genericEmailPublisher := NewGenericEmailQueuePublisher(ch, exchangeName)

	return PublisherManager{
		conn:                       conn,
		ch:                         ch,
		GenericEmailQueuePublisher: genericEmailPublisher,
	}
}

func (qcm *PublisherManager) Close() {
	err := qcm.ch.Close()
	failOnError(err, "Failed to close channel")

	err = qcm.conn.Close()
	failOnError(err, "Failed to close connection")
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
