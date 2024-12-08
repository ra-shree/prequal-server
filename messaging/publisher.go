package messaging

import (
	"context"
	"encoding/json"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var conn *amqp.Connection
var ch *amqp.Channel

// InitializePublisher sets up the connection and channel for the publisher.
func InitializePublisher() {
	var err error

	conn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err = conn.Channel()
	failOnError(err, "Failed to open a channel")
}

// PublishMessage publishes a message to the specified queue.
func PublishMessage(queueName string, message *Message) error {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Ensure the queue exists
	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Publish the message
	err = ch.PublishWithContext(ctx,
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        messageBytes,
		},
	)

	if err != nil {
		return err
	}

	log.Printf(" [x] Sent %s\n", string(messageBytes))
	return nil
}

// CleanupPublisher closes the channel and connection.
func CleanupPublisher() {
	if ch != nil {
		ch.Close()
	}
	if conn != nil {
		conn.Close()
	}
}
