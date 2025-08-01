package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/websocket"
)

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
            <-cfg.ready  //hold until filename set
            if len(cfg.filename) > 0 {
                // This should execute on the first pass to create the file or force rename
                filename := <-cfg.filename
                if _, err := os.Stat(filepath.Join(cfg.ftpdDir, filename)); os.IsExist(err) {
                    fmt.Printf("file %s exists. Enter a new name: ", filename)
                    cfg.filename <- filename
                    continue
                }
                file, err := os.Create(filepath.Join(cfg.ftpdDir, filename))
                if err != nil {
                    log.Printf("error writing file: %s\n", err)
                    continue
                }
                cfg.file = file
            }
            // if len(cfg.ready) == 0 {
            //     log.Print("unable to process binary data from websocket")
            //     continue
            // }
            cfg.writeFile(message)
            cfg.ready <- true
        }
    }
}

func (cfg *apiConfig) writeFile(buf []byte) error {
    cfg.file.Write(buf)
    return nil
}
