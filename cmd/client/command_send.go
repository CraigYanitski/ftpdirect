package main

import (
	"os"

	"github.com/gorilla/websocket"
)

func commandSend(cfg *apiConfig, arg string) error {
    file, err := os.Open(arg)
    if err != nil {
        return err
    }
    cfg.sendFile(file)
    return nil
}

func (cfg *apiConfig) sendFile(file *os.File) {
    buf := make([]byte, 1024)
    for {
        n, err := file.Read(buf)
        if err != nil {
            break
        }
        if cfg.internal {
            cfg.peerConn.Write(buf[:n])
        } else {
            cfg.ws.WriteMessage(websocket.BinaryMessage, buf[:n])
        }
    }
    return
}
