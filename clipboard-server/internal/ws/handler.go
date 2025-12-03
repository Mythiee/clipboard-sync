package ws

import (
    "fmt"
    "github.com/gorilla/websocket"
    "github.com/mythiee/clipboard-sync/clipboard-server/pkg/utils"
    "net/http"
    "time"
)

func Handler(w http.ResponseWriter, r *http.Request) {
    // upgrade HTTP to websocket
    conn, err := utils.WsUpgrader.Upgrade(w, r, nil)
    if err != nil {
        fmt.Println(err)
        return
    }

    defer func() {
        if err := conn.Close(); err != nil {
            fmt.Println(err)
        }
    }()
    fmt.Println("client connected")

    // heartbeat config
    pingInterval := 10 * time.Second
    pongWait := 30 * time.Second
    readErr := conn.SetReadDeadline(time.Now().Add(pongWait))
    if readErr != nil {
        fmt.Println(readErr)
        return
    }

    // pong handler
    conn.SetPongHandler(func(string) error {
        readErr = conn.SetReadDeadline(time.Now().Add(pongWait))
        if readErr != nil {
            return readErr
        }
        return nil
    })

    // send ping with timeout 1s
    go func() {
        for {
            time.Sleep(pingInterval)
            err := conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(1*time.Second))
            if err != nil {
                fmt.Println("send ping error: ", err)
                if err := conn.Close(); err != nil {
                    fmt.Println(err)
                }
                return
            }
        }
    }()

    for {
        msgType, msg, err := conn.ReadMessage()
        if err != nil {
            fmt.Println(err)
            break
        }

        // only text
        if msgType == websocket.TextMessage {
            fmt.Printf("Received message: %s\n", msg)
        }
    }

    fmt.Println("client disconnected")
}
