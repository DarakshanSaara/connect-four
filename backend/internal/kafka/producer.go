package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

type GameEvent struct {
	Type      string    `json:"type"` // game_start, move, game_end
	GameID    string    `json:"gameId"`
	PlayerID  string    `json:"playerId"`
	Username  string    `json:"username"`
	Data      string    `json:"data"`
	Timestamp time.Time `json:"timestamp"`
}

func NewProducer(brokers []string) *Producer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    "game-events",
		Balancer: &kafka.LeastBytes{},
	}

	return &Producer{writer: writer}
}

func (p *Producer) SendEvent(event GameEvent) error {
	event.Timestamp = time.Now()
	jsonData, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(event.GameID),
		Value: jsonData,
	}

	return p.writer.WriteMessages(context.Background(), msg)
}

func (p *Producer) Close() error {
	return p.writer.Close()
}