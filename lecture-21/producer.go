package main

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

func main() {

	// функциональная абстракция для реализации логирования в библиотеке kafka-go
	//    1) Для информирующих сообщений
	InfoLogger := kafka.LoggerFunc(func(msg string, args ...interface{}) {
		// тут вы вольны реализовывать логирование как вам надо
		log.Info().Msgf(msg, args...)
	})
	//    2) Для сообщений об ошибках
	ErrorLogger := kafka.LoggerFunc(func(msg string, args ...interface{}) {
		// тут вы вольны реализовывать логирование как вам надо
		log.Error().Msgf(msg, args...)
	})

	// kafkaWriter - instance структуры для записи данных в kafka. Библиотека kafka-go достаточно хорошо описана внутри.
	//    Особенно эта структура, если у вас есть вопросы и вы не нашли ответ тут - сначала посмотрите комментарии внутри.
	//    В этом примере указанны поля, которых предельно достаточно для работы с продюсером.
	kafkaWriter := &kafka.Writer{
		Addr: kafka.TCP("127.0.0.1:9092"), // тут все понятно - вы должны указать все адреса всего кластера

		Balancer: &kafka.LeastBytes{}, // реализация балансировки между партициями, по дефолту &kafka.RoundRobin{}.
		//
		RequiredAcks: kafka.RequireOne, // достаточно важное поле - нужны ли acknowledgements от kafka, при записи сообщений.
		//                                    является компромиссом. есть 3 случая:
		// RequireNone - fire-and-forget. Самый быстрый и самый опасный вариант.
		// RequireOne  - подтверждение записи идет только от локальной ноды, куда в этот момент писал наш продюсер.
		//                  Он медленнее варианта с  RequireNone, но гораздо быстрее RequireAll.
		//                  Компромисс: мы уверены, что запись прошла. Но Мы не можем гарантировать:
		//                      1) что в этот момент с локальной нодой ничего не случится.
		//                      2) т.к. kafka - это распределенная система хранения данных в виде очереди, могут быть разночтения.
		//                      3) есть небольшая вероятность при смете активной ноды, что консьюмеры (читатели) могут пропустить часть данных.
		// RequireAll  - самый долгий, но самый надежный способ сохранить данные на kafka.
		//                  Какие подводные камни есть:
		//                     1) Параметр зависит от настройки кластера. Если кластер - это 3 ноды обычно делают подтверждение 2-мя.
		//                           Вот тут первый камень. Продюсер будет ждать либо пока как минимум 2 ноды подтвердят запись,
		//                              либо отдаст ошибку по timeout и сделает rollback - удалит данные, если те смогли куда-то записаться
		//                     2) Если выставить 3 из 3-х. То при падении одной ноды вы никогда не сможете сохранить данные на кластер.

		BatchSize: 100, // параметр отвечает за то какими пачками (по количеству сообщений) мы будем стрелять в kafka.
		//                    В kafka всегда стреляют пачками сообщений.
		//                    Является одним из триггеров на отправку сообщений.

		BatchBytes: 1048576, // параметр отвечает за то какими пачками (по размеру батча) мы будем стрелять в kafka.
		//                         Является одним из триггеров на отправку сообщений.

		BatchTimeout: time.Second, // Отсечка. Если нужный батч не собрался до этого времени - отправляем что есть.
		//                               Является одним из триггеров на отправку сообщений.

		Compression: kafka.Gzip, // метода сжатья данных. Рекомендую использовать - Gzip. За все время ни одной проблемы не было выявлено.

		MaxAttempts: 50, // параметр указывает, сколько попыток надо сделать для отправки сообщения. Я выставляю 50.
		//                      !!!ВАЖНО!!! - не ориентируйтесь на данный параметр. Подтверждайте доставку самостоятельно. (реализация ниже)

		WriteTimeout: time.Second * 10, // timeout, на запись батча в kafka, после которой kafka отдаст ошибку.

		Async: false, // параметр позволяющий переключаться между асинхронным и синхронным вызовом метода WriteMessages (метод блокируется или нет).
		//                         !!!ВАЖНО!!! - если вы используете вариант с асинхронной работой продюсера.
		//                            1) у вас должен быть очень веский повод использовать именно его. Потому что он не гарантирует доставку.
		//                            2) нужно реализовать поле Completion - Completion func(messages []Message, err error) в отдельной горутине.
		// Из документации:
		/*
			// Setting this flag to true causes the WriteMessages method to never block.
			// It also means that errors are ignored since the caller will not receive
			// the returned value. Use this only if you don't care about guarantees of
			// whether the messages were written to Kafka.
			Async bool

			// An optional function called when the writer succeeds or fails the
			// delivery of messages to a Kafka partition. When writing the messages
			// fails, the `err` parameter will be non-nil.
			//
			// The messages that the Completion function is called with have their
			// topic, partition, offset, and time set based on the Produce responses
			// received from kafka. All messages passed to a call to the function have
			// been written to the same partition. The keys and values of messages are
			// referencing the original byte slices carried by messages in the calls to
			// WriteMessages.
			//
			// The function is called from goroutines started by the writer. Calls to
			// Close will block on the Completion function calls. When the Writer is
			// not writing asynchronously, the WriteMessages call will also block on
			// Completion function, which is a useful guarantee if the byte slices
			// for the message keys and values are intended to be reused after the
			// WriteMessages call returned.
			//
			// If a completion function panics, the program terminates because the
			// panic is not recovered by the writer and bubbles up to the top of the
			// goroutine's call stack.
			Completion func(messages []Message, err error)
		*/

		Logger: InfoLogger, // сюда присваиваем функцию которую реализовали немного выше - InfoLogger.

		ErrorLogger: ErrorLogger, // сюда присваиваем функцию которую реализовали немного выше - ErrorLogger.
		//                              Если у вас есть реализация InfoLogger, но нету ErrorLogger. Все ошибки будут писаться к InfoLogger.
		//                              Но наоборот эта схема не работает. Реализовав только ErrorLogger,
		//                                 вы будите получать только сообщения об ошибках, что удобно на продуктиве.
	}

	//т.к. это только пример пусть будет одно сообщение. У себя вы можете изменить это на цикл и миллионами сообщений.
	//   главное создать список заранее. И да можно писать в несколько топиков используя один массив с сообщениями
	msgs := []kafka.Message{
		{
			Topic: "test",
			Value: []byte("testValue1"),
		},
		{
			Topic: "test",
			Value: []byte("testValue2"),
		},
		{
			Topic: "test",
			Value: []byte("testValue3"),
		}, {
			Topic: "test",
			Value: []byte("testValue4"),
		}, {
			Topic: "test",
			Value: []byte("testValue5"),
		}, {
			Topic: "test",
			Value: []byte("testValue6"),
		},
		{
			Topic: "test",
			Value: []byte("testValue7"),
		},
	}

	// Реализация гарантированной доставки сообщений в kafka.
	// Тут много зависит от вас. Т.к. вы решаете какое значение будет у maxAttempts.
	const maxAttempts = 100

	connectErrorsCount := 0 // count для подсчета неудачных попыток с непредвиденными ошибками.
	//                                 Такие ошибки обычно появляются когда kafka cluster недоступен.

LOOP:
	for {
		errCount := 0            // Информативный count для лога. Сколько сообщений не отправилось
		succCountIfErrOccur := 0 // Информативный count для лога. Сколько сообщений отправилось успешно

		switch err := kafkaWriter.WriteMessages(context.Background(), msgs...).(type) { // попытка отправки сообщений в kafka
		case nil: // случай, когда все сообщения успешно доставлены
			log.Info().Msgf("All messages success send to kafka. Count: %d.", len(msgs))
			break LOOP
		case kafka.WriteErrors: // Особенный тип ошибки помогающий нам выловить сообщения который мы не смогли отослать по какой-то причине.
			//                            тип этой ошибки -> []err. И логика такая, что он возвращает массив ошибок идентичным
			//                               размером вашего массива с сообщениями для отправки (msgs). После чего вам нужно пройти циклом
			//                               по вашему изначальному массиву с сообщениями (msgs).
			//                               Далее вы получаете индекс своего сообщения (for -> i <- := range msgs)
			//                               и используя этот индекс вы идете в возвращенную ошибку (err. Напоминаю что это массив)
			//                               проверяете err[i] на nil, если он не nil - нужно записать сообщение под этим индексом (msgs[i]) в новый, заранее подготовленный, массив (failedMsgs).
			//                               После того как вы прошли весь свой массив, вы должны переопределить переменные msgs = failedMsgs, и повторить отправку сообщений (msgs) снова.
			//                               Единственная возможность покинуть этот цикл отправить все сообщения до последнего. Либо если произойдет 100 непредвиденных ошибок.
			//                               !!!ВАЖНО!!! - kafka.WriteErrors - является нормой, поэтому connectErrorsCount не инкриминируется.

			var failedMsgs []kafka.Message // временный массив для сбора сообщений, которые не смогли отправиться в kafka.
			for i := range msgs {          // 1) проходимся по вашему списку с сообщениями на отправку.
				if err[i] != nil { // 2) используя i проверяем наличие ошибки в err
					failedMsgs = append(failedMsgs, msgs[i]) // 3) Если есть, то мы знаем что msgs[i] не было доставлено.
					//                                          Его нужно куда-то временно записать, что бы потом переотправить.
					errCount++ // увеличиваем count для логирования
				} else {
					succCountIfErrOccur++ // увеличиваем count для логирования
				}
			}
			log.Info().Msgf("failed messages info: Failed: %d, Success: %d, Total: %d.", errCount, succCountIfErrOccur, len(msgs))
			msgs = failedMsgs // 4) переопределяем переменную msgs, чтобы повторно отправить недоставленные в предыдущей итерации сообщения.
		default: // Если возникла большая непредвиденная ошибка.
			// тут желательно реализовать !!!backoff!!!
			if connectErrorsCount > maxAttempts {
				log.Fatal().Err(err).Msgf("connectErrorsCount > 100") // это может возникнуть, например если kafka кластер не доступен.
			}
			log.Error().Msgf("WriteMessagesError: %s.\n", err) // пишем в лог ошибку и пробуем еще раз до maxAttempts
			connectErrorsCount++                               // инкриминируем, чтобы знать количество попыток.
		}
	}
}
