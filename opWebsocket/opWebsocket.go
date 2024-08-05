package opWebsocket

import (
	"Stock/opConfig"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type User struct {
	Username string
	Conn     *websocket.Conn
	Stocks   string
}

var (
	clients      = make(map[*websocket.Conn]*User)
	clientsMutex = &sync.Mutex{}
	upgrader     = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	defer func() {
		clientsMutex.Lock()
		delete(clients, conn)
		clientsMutex.Unlock()
		conn.Close()
	}()

	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Printf("Error reading message: %v", err)
		return
	}

	user := string(message)

	clientsMutex.Lock()
	clients[conn] = &User{Conn: conn, Username: user}
	clientsMutex.Unlock()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func GetUsersStocks() map[*websocket.Conn]*User {
	usersStocks := make(map[*websocket.Conn]*User)
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for conn, user := range clients {
		user.Stocks = opConfig.GetUserStocks(user.Username)
		usersStocks[conn] = user
	}

	return usersStocks
}

func BroadcastMessage(conn *websocket.Conn, message []byte) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	err := conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		log.Printf("Error broadcasting message: %v", err)
		conn.Close()
		delete(clients, conn)
	}
}
