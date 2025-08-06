package main

import (
	"fmt"

	"github.com/gorilla/websocket"
)

func commandExit(cfg *apiConfig, arg string) error {
    fmt.Println("Closing FTP direct... Goodbye!")
    cfg.ws.WriteMessage(websocket.CloseMessage, []byte{})
    cfg.ctxCancel()
    return nil
}
