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

var errExists = errors.New("File already exists")

func wsListener(cfg *apiConfig) {
    for {
        msgType, message, err := cfg.ws.ReadMessage()
        if err != nil {
            log.Printf("error reading from websocket: %s\n", err)
            return
        }
        switch msgType {
        case websocket.TextMessage:
            fmt.Print("\033[G\033[K")
            msgArgs := strings.Fields(string(message))
            log.Printf("received: %s\n", message)
            if msgArgs[0] == "Connected" {
                cfg.peer = msgArgs[len(msgArgs)-1]
                fmt.Printf("    -> connected to room %s\n", cfg.peer)
            } else if msgArgs[0] == "Disconnecting" {
                cfg.peer = ""
                fmt.Printf("    -> disconnected from %s\n", msgArgs[len(msgArgs)-1])
            } else if msgArgs[0] == "Sending" {
                filename := strings.Join(msgArgs[2:], " ")
                fmt.Printf("Receive file %s. (Optional) enter an alternative name: ", filename)
                cfg.filename <- filename
                // cfg.ready <- false
                // fmt.Println("Processing...")
            } else if string(message) == "Done sending file" {
                cfg.file.Close()
                cfg.file = nil
                <-cfg.ready
                log.Println("\nClosing file")
            }
        case websocket.BinaryMessage:
            // fmt.Printf("%x", message)
            // This should execute on the first pass when a file is sent to either create the file or force rename
            for len(cfg.filename) > 0 {
                <-cfg.ready  //hold until filename set
                filename := <-cfg.filename
                err = cfg.createFile(filename)
                if err == errExists {
                    if !os.IsNotExist(err) {
                        fmt.Printf("File %s exists. Enter a new name: ", filename)
                    }
                    cfg.filename <- filename
                    continue
                }
                cfg.ready <- true
                break
            }
            <-cfg.ready
            cfg.writeFile(message)
            cfg.ready <- true
        case websocket.CloseMessage:
            fmt.Print("\033[G\033[K")
            log.Printf("received: %s\n", message)
            return
        }
    }
}

func (cfg *apiConfig) createFile(filename string) error {
    if _, err := os.Stat(filepath.Join(cfg.ftpdDir, filename)); err == nil {
        return errExists
    }
    file, err := os.Create(filepath.Join(cfg.ftpdDir, filename))
    if err != nil {
        log.Printf("error writing file: %s\n", err)
        return err
    }
    cfg.file = file
    return nil
}

func (cfg *apiConfig) writeFile(buf []byte) error {
    cfg.file.Write(buf)
    return nil
}

