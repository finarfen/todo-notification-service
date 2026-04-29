package main

import (
	"fmt"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	host := getEnv("RABBITMQ_HOST", "localhost")
	port := getEnv("RABBITMQ_PORT", "5672")
	user := getEnv("RABBITMQ_USER", "guest")
	pass := getEnv("RABBITMQ_PASS", "guest")

	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, pass, host, port)

	var conn *amqp.Connection
	var err error

	for i := 0; i < 10; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		log.Printf("RabbitMQ недоступен, попытка %d/10...", i+1)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		log.Fatal("Не удалось подключиться к RabbitMQ:", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Ошибка открытия канала:", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"notifications", true, false, false, false, nil,
	)
	if err != nil {
		log.Fatal("Ошибка создания очереди:", err)
	}

	msgs, err := ch.Consume(
		q.Name, "", true, false, false, false, nil,
	)
	if err != nil {
		log.Fatal("Ошибка подписки на очередь:", err)
	}

	fmt.Println("Notification service запущен, ожидаю сообщения...")

	forever := make(chan bool)
	go func() {
		for msg := range msgs {
			fmt.Printf("Получено уведомление: %s\n", msg.Body)
		}
	}()
	<-forever
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
