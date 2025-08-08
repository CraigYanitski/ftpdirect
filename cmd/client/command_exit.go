package main

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

func commandExit(cfg *apiConfig, arg string) error {
    fmt.Println("Closing FTP direct... Goodbye!")
    cfg.ws.WriteMessage(websocket.TextMessage, []byte("Close"))
    time.Sleep(2*time.Second)
    cfg.ctxCancel()
    return nil
}
