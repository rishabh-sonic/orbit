package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rishabh-sonic/orbit/internal/auth"
	"github.com/rishabh-sonic/orbit/internal/middleware"
	"github.com/rishabh-sonic/orbit/internal/notification"
	"github.com/rishabh-sonic/orbit/internal/wsHub"
	mq "github.com/rishabh-sonic/orbit/pkg/rabbitmq"
	"github.com/rishabh-sonic/orbit/pkg/config"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}

	jwtSvc := auth.NewJWTService(cfg.JWTSecret, cfg.JWTResetSecret, cfg.JWTExpiration)
	hub := wsHub.New()

	mqClient, err := mq.New(cfg)
	if err != nil {
		slog.Error("rabbitmq connect", "err", err)
		os.Exit(1)
	}
	defer mqClient.Close()

	// Start consuming from all 16 shard queues
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	for _, queueName := range mq.ShardQueues() {
		deliveries, err := mqClient.Consume(queueName)
		if err != nil {
			slog.Error("consume queue", "queue", queueName, "err", err)
			continue
		}
		wg.Add(1)
		go func(queue string, msgs <-chan amqp.Delivery) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case msg, ok := <-msgs:
					if !ok {
						return
					}
					dispatch(hub, msg.Body)
					msg.Ack(false)
				}
			}
		}(queueName, deliveries)
	}

	// HTTP server for WS connections
	r := chi.NewRouter()
	r.Use(middleware.Authenticate(jwtSvc))

	r.Get("/api/ws", func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Error("ws upgrade", "err", err)
			return
		}

		hub.Register(claims.UserID, conn)
		defer func() {
			hub.Unregister(claims.UserID, conn)
			conn.Close()
		}()

		// Keep connection alive, read pings
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	})

	addr := fmt.Sprintf(":%s", cfg.WSPort)
	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		slog.Info("ws service starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("ws server", "err", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancel()
	wg.Wait()
	srv.Shutdown(context.Background())
	slog.Info("ws service stopped")
}

func dispatch(hub *wsHub.Hub, body []byte) {
	var payload notification.WirePayload
	if err := json.Unmarshal(body, &payload); err != nil {
		slog.Error("dispatch unmarshal", "err", err)
		return
	}
	userID := payload.UserID
	if userID == uuid.Nil {
		return
	}
	hub.Send(userID, body)
}
