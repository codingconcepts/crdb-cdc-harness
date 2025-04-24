package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/codingconcepts/crdb-cdc-harness/pkg/models"
	"github.com/codingconcepts/crdb-cdc-harness/pkg/random"
	"github.com/codingconcepts/crdb-cdc-harness/pkg/views"
	"github.com/codingconcepts/env"
	"github.com/codingconcepts/goutil/duration"
	"github.com/codingconcepts/goutil/safemap"
	"github.com/codingconcepts/ring"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/lo"
	"github.com/segmentio/kafka-go"
	"github.com/tidwall/gjson"
)

var (
	version       string
	writtenCount  uint64
	writingActive = true
	writtenIDs    = safemap.New[string, time.Time]()
)

func main() {
	var e models.EnvironmentVariables

	flag.StringVar(&e.DatabaseURL, "database-url", "", "database connection string")
	flag.StringVar(&e.KafkaURL, "kafka-url", "", "url to the kafka broker")
	flag.StringVar(&e.KafkaTopic, "kafka-topic", "", "name of the kafka topic")
	flag.StringVar(&e.KafkaMessageIDPath, "kafka-message-id-path", "", "dot notation path to the id in the kafka message")
	flag.DurationVar(&e.WriteFrequency, "write-frequency", time.Second*1, "database write frequency")
	flag.IntVar(&e.LatencyPoolSize, "latency-pool-size", 100, "average latency sample size")
	flag.StringVar(&e.WriteStatement, "write-statement", "", "database write statement to use")
	flag.Var(&e.Args, "arg", "argument generator")
	showVersion := flag.Bool("version", false, "display the application version")

	flag.Parse()

	if *showVersion {
		log.Printf("cdch version: %s", version)
		return
	}

	// Override settings with values from the environment if provided.
	if err := env.Set(&e); err != nil {
		log.Fatalf("error setting environment variables: %v", err)
	}

	// Fail if any required fields are missing.
	if e.DatabaseURL == "" || e.WriteStatement == "" || e.KafkaURL == "" || e.KafkaTopic == "" {
		flag.Usage()
		os.Exit(2)
	}

	db, err := pgxpool.New(context.Background(), e.DatabaseURL)
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	defer db.Close()

	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{e.KafkaURL},
		GroupID:     uuid.NewString(),
		Topic:       e.KafkaTopic,
		StartOffset: kafka.LastOffset,
	})

	messages := make(chan models.KafkaMessage, 1000)
	status := make(chan bool, 10)

	go readKafkaMessages(e, kafkaReader, messages)
	go manageDatabaseWrites(e, db)

	p := tea.NewProgram(views.NewModel(messages, status))

	go listen(e, messages, status, p)

	fmt.Print("\033[H\033[2J")
	if _, err := p.Run(); err != nil {
		log.Fatalf("error running UI: %v", err)
	}
}

func listen(env models.EnvironmentVariables, messages <-chan models.KafkaMessage, status <-chan bool, p *tea.Program) {
	latencies := ring.New[time.Duration](env.LatencyPoolSize)
	maxOffset := int64(-1)
	printTick := time.Tick(time.Second)

	var lastMessageReceivedTime time.Time

	for {
		select {

		// Fires on every status change.
		case status := <-status:
			writingActive = status

		// Fires on every new message received from Kafka.
		case m := <-messages:
			// Update the last message time
			lastMessageReceivedTime = time.Now()

			// Only process each message once.
			if m.Offset <= int64(maxOffset) {
				continue
			}

			// Fetch time from the writtenIDs map and skip if ID not found.
			writtenTime, ok := writtenIDs.Get(m.RowID)
			if !ok {
				continue
			}

			latencies.Add(time.Since(writtenTime))
			maxOffset = m.Offset

			// Remove the read item.
			writtenIDs.Delete(m.RowID)

		// Fires every second and updates the terminal
		case <-printTick:
			var avgLatency time.Duration

			// If we've not received a Kafka message for more than 5 seconds, derive
			// the latency from the time we last saw a message.
			noMessagesFor := time.Since(lastMessageReceivedTime)
			if noMessagesFor >= time.Second*5 && writingActive {
				avgLatency = noMessagesFor
			} else {
				latencySlice := latencies.Slice()
				if len(latencySlice) > 0 {
					avgLatency = lo.Sum(latencySlice) / time.Duration(len(latencySlice))
				}
			}

			p.Send(views.LatencyUpdateMsg{
				AvgLatency: duration.Round(avgLatency, time.Millisecond),
			})

			p.Send(views.UpdateStatsMsg{
				CountUnread:  writtenIDs.Len(),
				CountWritten: atomic.LoadUint64(&writtenCount),
			})
		}
	}
}

func readKafkaMessages(env models.EnvironmentVariables, kafkaReader *kafka.Reader, messages chan<- models.KafkaMessage) error {
	for {
		msg, err := kafkaReader.FetchMessage(context.Background())
		if err != nil {
			log.Printf("error reading message: %v", err)
			continue
		}

		kafkaMessage := models.KafkaMessage{
			Message:      msg,
			ReceivedTime: time.Now(),
			RowID:        gjson.GetBytes(msg.Value, env.KafkaMessageIDPath).String(),
		}

		// Let other channel know about this messages.
		messages <- kafkaMessage

		if err = kafkaReader.CommitMessages(context.Background(), msg); err != nil {
			log.Printf("error committing message: %v", err)
		}
	}
}

func manageDatabaseWrites(env models.EnvironmentVariables, db *pgxpool.Pool) {
	ticker := time.Tick(env.WriteFrequency)

	for range ticker {
		if !writingActive {
			continue
		}

		args, err := random.GenerateArgValues(env.Args)
		if err != nil {
			log.Printf("error generating args values: %v", err)
			continue
		}

		row := db.QueryRow(context.Background(), env.WriteStatement, args...)

		var id string
		if err = row.Scan(&id); err != nil {
			log.Printf("error scanning written row: %v", err)
			continue
		}

		// Add written row
		writtenIDs.Set(id, time.Now())
		atomic.AddUint64(&writtenCount, 1)
	}
}
