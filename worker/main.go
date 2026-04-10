package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	kingpin "github.com/alecthomas/kingpin/v2"

	"github.com/IBM/sarama"
)

var (
	brokerList = kingpin.Flag("brokerList", "List of brokers to connect").Default(getEnv("KAFKA_BROKERS", "kafka:29092")).Strings()
	topic      = kingpin.Flag("topic", "Topic name").Default(getEnv("KAFKA_TOPIC", "votes")).String()
	group      = kingpin.Flag("group", "Kafka consumer group").Default(getEnv("KAFKA_GROUP", "voting-group")).String()
)

const (
	defaultHost     = "postgres"
	defaultPort     = "5432"
	defaultUser     = "okteto"
	defaultPassword = "okteto"
	defaultDBName   = "votes"
)

func main() {
	kingpin.Parse()

	db := openDatabase()
	defer db.Close()

	pingDatabase(db)

	createTableStmt := `CREATE TABLE IF NOT EXISTS votes (id VARCHAR(255) NOT NULL UNIQUE, vote VARCHAR(255) NOT NULL)`
	if _, err := db.Exec(createTableStmt); err != nil {
		log.Panic(err)
	}

	consumerGroup := getKafkaConsumerGroup()
	defer consumerGroup.Close()

	processed := &atomic.Uint64{}
	handler := &voteConsumerGroupHandler{db: db, processed: processed}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-signals
		fmt.Println("Interrupt is detected")
		cancel()
	}()

	go func() {
		for {
			if err := consumerGroup.Consume(ctx, []string{*topic}, handler); err != nil {
				log.Printf("consumer group error: %v", err)
				time.Sleep(1 * time.Second)
			}

			if ctx.Err() != nil {
				return
			}
		}
	}()

	<-ctx.Done()

	log.Println("Processed", processed.Load(), "messages")
}

type voteConsumerGroupHandler struct {
	db        *sql.DB
	processed *atomic.Uint64
}

func (h *voteConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *voteConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *voteConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		voterID := string(msg.Key)
		voteValue := string(msg.Value)

		if voterID == "" {
			log.Printf("Skipping message at offset %d because key is empty", msg.Offset)
			session.MarkMessage(msg, "empty-key")
			continue
		}

		fmt.Printf("Received message: user %s vote %s\n", voterID, voteValue)

		insertDynStmt := `insert into "votes"("id", "vote") values($1, $2) on conflict(id) do update set vote = $2`
		if _, err := h.db.Exec(insertDynStmt, voterID, voteValue); err != nil {
			log.Printf("Error persisting vote for voter %s: %v", voterID, err)
			session.MarkMessage(msg, "db-error")
			continue
		}

		h.processed.Add(1)
		session.MarkMessage(msg, "processed")
	}

	return nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func openDatabase() *sql.DB {
	host := getEnv("POSTGRES_HOST", defaultHost)
	port := getEnv("POSTGRES_PORT", defaultPort)
	user := getEnv("POSTGRES_USER", defaultUser)
	password := getEnv("POSTGRES_PASSWORD", defaultPassword)
	dbname := getEnv("POSTGRES_DB", defaultDBName)

	psqlconn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	for {
		db, err := sql.Open("postgres", psqlconn)
		if err == nil {
			return db
		}
		time.Sleep(1 * time.Second)
	}
}

func pingDatabase(db *sql.DB) {
	fmt.Println("Waiting for postgresql...")
	for {
		if err := db.Ping(); err == nil {
			fmt.Println("Postgresql connected!")
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func getKafkaConsumerGroup() sarama.ConsumerGroup {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Return.Errors = true
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	brokers := *brokerList
	if len(brokers) == 1 && strings.Contains(brokers[0], ",") {
		brokers = strings.Split(brokers[0], ",")
	}

	fmt.Println("Waiting for kafka...")
	for {
		consumerGroup, err := sarama.NewConsumerGroup(brokers, *group, config)
		if err == nil {
			fmt.Println("Kafka connected!")
			return consumerGroup
		}
		time.Sleep(1 * time.Second)
	}
}
