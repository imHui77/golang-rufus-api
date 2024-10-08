package mqtt

import (
	"log"
	"math"
	"time"

	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	conn         *amqp.Connection
	ch           *amqp.Channel
	isConnected  bool
	reconnectCh  chan bool
	maxRetries   int
	retryTimeout time.Duration
}

func InitRabbitMQ(url string) (*RabbitMQ, error) {
	rmq := &RabbitMQ{
		reconnectCh:  make(chan bool),
		maxRetries:   5,
		retryTimeout: 5 * time.Second,
	}

	err := rmq.connect(url)
	if err != nil {
		return nil, err
	}

	go rmq.handleReconnect(url)

	return rmq, nil
}

func (r *RabbitMQ) connect(url string) error {
	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	r.conn = conn
	r.ch = ch
	r.isConnected = true

	go func() {
		<-r.conn.NotifyClose(make(chan *amqp.Error))
		r.isConnected = false
		r.reconnectCh <- true
	}()

	return nil
}

func (r *RabbitMQ) handleReconnect(url string) {
	for range r.reconnectCh {
		if !r.isConnected {
			log.Println("Attempting to reconnect to RabbitMQ...")

			for i := 0; i < r.maxRetries; i++ {
				err := r.connect(url)
				if err == nil {
					log.Println("Successfully reconnected to RabbitMQ")
					break
				}

				log.Printf("Failed to reconnect: %v. Retrying in %v...", err, r.retryTimeout)
				time.Sleep(r.retryTimeout)
				r.retryTimeout *= 2 // 指數退避
				if r.retryTimeout > 1*time.Minute {
					r.retryTimeout = 1 * time.Minute // 最大重試間隔為1分鐘
				}
			}

			if !r.isConnected {
				log.Println("Failed to reconnect after maximum retries. Exiting...")
				// 在這裡可以添加額外的錯誤處理邏輯，例如發送警報或退出應用程序
			}
		}
	}
}

func (r *RabbitMQ) PublishMessage(queue string, body []byte) error {
	if !r.isConnected {
		return amqp.ErrClosed
	}

	q, err := r.ch.QueueDeclare(
		queue, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	return r.ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}

func (r *RabbitMQ) ConsumeMessages(queueName string, handler func([]byte) error) error {
	if !r.isConnected {
		return amqp.ErrClosed
	}

	q, err := r.ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return err
	}

	msgs, err := r.ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			err := handler(msg.Body)
			if err != nil {
				log.Printf("Error processing message: %v", err)
			}
		}
	}()

	log.Printf("Started consuming messages from queue: %s", queueName)
	return nil
}

func (r *RabbitMQ) Close() {
	if r.ch != nil {
		r.ch.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}