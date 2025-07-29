package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	// "strings"

	"github.com/gorilla/websocket"
)

func commandConnect(cfg *apiConfig, arg string) error {
    if cfg.internal {
        if arg == "" {
            err := errors.New("Must identify TCP IP address for internal connections")
            log.Println(err)
            return err
        }
        conn, err := net.Dial("tcp", arg)
        if err != nil {
            return err
        }
        cfg.peerConn = conn.(*net.TCPConn)
    } else {
        var cmd string
        if len(arg) > 0 {
            cmd = fmt.Sprintf("Connect %s", arg)
        } else {
            cmd = "Connect"
        }
        cfg.ws.WriteMessage(
            websocket.TextMessage,
            []byte(cmd),
        )
        // _, msgByte, err := cfg.ws.ReadMessage()
        // if err != nil {
        //     return err
        // }
        // msg := string(msgByte)
        // log.Println(msg)
        // if strings.HasPrefix(msg, "Connected") {
        //     fields := strings.Fields(msg)
        //     cfg.peer = fields[len(fields)-1]
        //     log.Printf("connected to room %s\n", cfg.peer)
        // } else {
        //     return errors.New(msg)
        // }
    }
    return nil
}
