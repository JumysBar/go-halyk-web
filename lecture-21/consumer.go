package main

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {

	// функциональная абстракция для реализации логирования в библиотеке kafka-go
	//    1) Для информирующих сообщений
	Info := kafka.LoggerFunc(func(msg string, args ...interface{}) {
		// тут вы вольны реализовывать логирование как вам надо
		log.Printf(msg, args...)
	})
	//    2) Для сообщений об ошибках
	Error := kafka.LoggerFunc(func(msg string, args ...interface{}) {
		// тут вы вольны реализовывать логирование как вам надо
		log.Printf(msg, args...)
	})

	// instance структуры kafka.Reader
	// этих полей более чем достаточно для работы с kafka.Reader
	kafkaR := kafka.NewReader(kafka.ReaderConfig{
		Brokers:       []string{"localhost:9092"}, // указываем все адеса нашего кластера
		GroupID:       "youGroupId",               // указываем group id, для консьюмера
		Topic:         "test",                     // наименование топика
		MinBytes:      10e3,                       // параметр отвечает за то насколько маленькие сообщения нужно читаться при вызове fetch
		MaxBytes:      10e6,                       // по аналогии с предыдущем параметром
		QueueCapacity: 1000,                       // ёмкость внутренней очереди. При инициации kafka.Reader создаётся отдельная горутина,
		//                                      которая читает kafka сообщения. И когда вы вызываете fetch, вы получаете сообщение не запросом в kafka,
		//                                      а из внутренней очереди уже прочитанных сообщений.
		StartOffset: kafka.FirstOffset, // этим параметром вы сообщаете откуда начать читать топик.
		//                                  Есть 2 параметра LastOffset и FirstOffset.
		//                                  Параметр отвечает на вопрос: "При создании нового консьюмера, надо начать читать от самого первого имеющегося сообщения или от самого последнего?"
		Logger:      Info,
		ErrorLogger: Error,
	})

	var batch []kafka.Message // место куда будут складываться все сообщения

	// цикл для наполнения batch сообщениями из kafka
	for i := 0; i < 5; i++ {
		msg, err := fetch(kafkaR) //метод описан немного ниже
		switch err {
		case nil:
		case context.DeadlineExceeded:
			log.Printf("warn deadline: %s", err)
			break
		default:
			log.Fatalf("default: %s", err)
		}
		batch = append(batch, msg)
	}

	// некий процесс над прочтенными сообщениями ДО коммита
	for _, msg := range batch {
		log.Printf("Your message is %s. Topic: %s, Partition: %d, Offset: %d",
			string(msg.Value), msg.Topic, msg.Partition, msg.Offset)
	}

	// непосредственно коммит ПОСЛЕ некого процесса над сообщениями
	if err := kafkaR.CommitMessages(context.Background(), batch...); err != nil {
		log.Fatalf("commit: %s", err)
	}

	log.Printf("all msgs were readed and commited")
}

// метод который позволит читать сообщения из kafka без commit
func fetch(reader *kafka.Reader) (kafka.Message, error) {
	// контекст нужен, если вы не хотите ждать вечность ваше сообщение.
	// Т.к. метод блокируется либо до получения сообщения, либо timeout по средствам context.WithTimeout
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*30)
	defer cancelFunc() //обязательно вызываем cancel функцию, чтобы нигде не было утечки памяти

	//довольно простой метод для получения одного сообщения из kafka
	msg, err := reader.FetchMessage(ctx)
	if err != nil {
		return kafka.Message{}, err
	}
	return msg, nil
}
