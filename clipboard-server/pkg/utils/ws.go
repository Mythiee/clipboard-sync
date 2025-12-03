package utils

import (
	"github.com/gorilla/websocket"
	"net/http"
)

var WsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// allow CORS
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
