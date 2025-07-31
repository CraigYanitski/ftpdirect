package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/gorilla/websocket"
)

func commandSend(cfg *apiConfig, arg string) error {
    if arg == "" {
        return errors.New("Must provide argument to `commandSend`")
    }
    cfg.ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Sending file %s\n", arg)))
    file, err := os.Open(arg)
    if err != nil {
        return err
    }
    cfg.sendFile(file)
    cfg.ws.WriteMessage(websocket.TextMessage, []byte("Done sending file"))
    return nil
}

func (cfg *apiConfig) sendFile(file *os.File) {
    /* Send file to appropriate connection */
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
