package main

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type apiConfig struct {
    ws        *websocket.Conn
    tcpConn   *net.Conn
    tcpAddr   string
    peer      string
    peerConn  *net.Conn
}

func startTCPServer() (string, error) {
    listener, err := net.Listen("tcp", ":0")
    if err != nil {
        return "", err
    }
    port := listener.Addr().(*net.TCPAddr).Port
    go func() {
        for {
            conn, err := listener.Accept()
            if err != nil{
                continue
            }
            go handleIncoming(conn)
        }
    }()
    return fmt.Sprintf("%s:%d", getLocalIP(), port), nil
}

func getLocalIP() string {
    conn, _ := net.Dial("udp", "8.8.8.8:80")
    defer conn.Close()
    return strings.Split(conn.LocalAddr().String(), ":")[0]
}

func handleIncoming(conn net.Conn) {
    defer conn.Close()
    buf := make([]byte, 1024)
    for {
        n, err := conn.Read(buf)
        if err != nil {
            break
        }
        fmt.Printf("%x", buf[:n])
    }
    return
}

func main() {
    interrupt := make(chan os.Signal, 1)
    signal.Notify(interrupt, os.Interrupt)

    u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/connect"}

    conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
    if err != nil {
        log.Fatalln(err)
    }
    defer conn.Close()

    tcpAddr, err := startTCPServer()
    if err != nil {
        log.Println("error starting TCP server")
        return
    }
    log.Printf("TCP listening on %s\n", tcpAddr)

    cfg := &apiConfig{
        ws: conn,
        tcpAddr: tcpAddr,
    }
    cfg.ws.WriteMessage(websocket.TextMessage, []byte(tcpAddr))

    startRepl(cfg)

    if len(os.Args) > 1 {
        peerAddr := os.Args[1]
        fmt.Printf("discovered peer: %s\n", peerAddr)

        tcpConn, err := net.Dial("tcp", peerAddr)
        if err != nil {
            log.Println(err)
        }
        defer tcpConn.Close()

        cfg.tcpConn = &tcpConn

        filename := "README.md"
        file, _ := os.Open(filename)
        defer file.Close()

        buf := make([]byte, 1024)
        for {
            n, err := file.Read(buf)
            if err != nil {
                break
            }

            (*cfg.tcpConn).Write(buf[:n])
        }
        log.Printf("sent file %s successfully to %s\n", filename, cfg.tcpAddr)
    }

    done := make(chan struct{})

    go func() {
        defer close(done)
        for {
            _, message, err := conn.ReadMessage()
            if err != nil {
                log.Println("error reading from websocket")
                return
            }
            log.Printf("received: %s\n", message)
        }
    }()

    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-done:
            return
        // case t := <-ticker.C:
        //     err := conn.WriteMessage(websocket.TextMessage, []byte(t.String()))
        //     if err != nil {
        //         log.Println("error writing time to websocket")
        //         return
        //     }
        case <-interrupt:
            log.Println("interrupt...")
            err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
            if err != nil {
                log.Println("Error writing close message")
                return
            }
            time.After(time.Second)
            select {
            case <-done:
            case <-time.After(time.Second):
            }
            return
        }
    }
}
