package ws

import (
	"github.com/gorilla/websocket"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, check frontend URL!
	},
}

func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	hub.Register(userID, conn)

	// 3. Keep connection alive & listen for disconnect
	go func() {
		defer hub.Unregister(userID)
		for {
			// We don't expect messages FROM the client in this setup,
			// but we must read to detect disconnection.
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}
