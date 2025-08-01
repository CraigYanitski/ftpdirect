package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/websocket"
)

func commandSend(cfg *apiConfig, arg string) error {
    if arg == "" {
        return errors.New("Must provide argument to `commandSend`")
    }
    cfg.ws.WriteMessage(
        websocket.TextMessage, 
        []byte(fmt.Sprintf("roomID %s Sending file %s\n", cfg.peer, arg)),
    )
    var filename = ""
    if strings.HasPrefix(arg, "/") {
        filename = arg
    } else {
        cwd, _ := os.Getwd()
        filename = filepath.Join(cwd, arg)
    }
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    cfg.sendFile(file)
    cfg.ws.WriteMessage(
        websocket.TextMessage, 
        []byte(fmt.Sprintf("roomID %s Done sending file", cfg.peer)),
    )
    log.Printf("sent file %s\n", filename)
    return nil
}

func (cfg *apiConfig) sendFile(file *os.File) {
    /* Send file to appropriate connection */
    buf := make([]byte, 1024)
    // log.Print("")
    for {
        n, err := file.Read(buf)
        if err != nil {
            fmt.Printf("\nerror reading file: %s\n", err)
            break
        }
        // fmt.Printf("%x", buf[:n])
        if cfg.internal {
            cfg.peerConn.Write(buf[:n])
        } else {
            cfg.ws.WriteMessage(websocket.BinaryMessage, buf[:n])
        }
    }
    return
}
