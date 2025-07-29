package main

import (
	"fmt"

	"github.com/gorilla/websocket"
)


func commandDisconnect(cfg *apiConfig, arg string) error {
    if !cfg.internal {
        cfg.ws.WriteMessage(
            websocket.TextMessage,
            []byte(fmt.Sprintf("Disconnect %s\n", cfg.peer)),
        )
    }
    cfg.peer = ""
    return nil
}
