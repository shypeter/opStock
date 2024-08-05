package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"Stock/opRedis"
	"Stock/opStock"
	"Stock/opWebsocket"

	"github.com/go-redis/redis/v8"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	userTicker := time.NewTicker(2 * time.Second)
	defer userTicker.Stop()

	rdb, err := opRedis.NewClient()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer rdb.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 設置 websocket路由
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "opStock.html")
	})
	http.HandleFunc("/ws", opWebsocket.HandleWebSocket)
	// 啟動 http server
	go func() {
		log.Println("Starting WebSocket server on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	var wg sync.WaitGroup

	for {
		select {
		case <-ticker.C:
			usersStocks := opWebsocket.GetUsersStocks()
			for _, user := range usersStocks {
				wg.Add(1)
				//log.Printf("get user %s stocks", user)
				go func(user *opWebsocket.User, rdb *redis.Client) {
					defer wg.Done()
					err := opStock.CallAPI(ctx, user.Stocks, rdb)
					if err != nil {
						log.Printf("Error calling API for user %s: %v", user.Username, err)
					}
				}(user, rdb)
			}
		case <-userTicker.C:
			usersStocks := opWebsocket.GetUsersStocks()
			for conn, user := range usersStocks {
				stocksInfo := opRedis.GetStocksInfo(ctx, rdb, user.Stocks)
				opWebsocket.BroadcastMessage(conn, []byte(stocksInfo))
			}
		case <-sigChan:
			log.Println("Received shutdown signal. Gracefully shutting down...")
			cancel()
			wg.Wait()
			log.Println("All goroutines have finished. Exiting.")
			return
		}
	}
}
