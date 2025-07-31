package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

type apiConfig struct {
    ctx        context.Context
    ftpdDir    string
    ws         *websocket.Conn
    tcpConn    *net.TCPConn
    tcpAddr    string
    peer       string
    peerConn   *net.TCPConn
    file       *os.File
    internal   bool
    filename   chan string
    ready      chan bool
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
    godotenv.Load()
    ftpdScheme := os.Getenv("FTPD_SCHEME")
    ftpdUrl := os.Getenv("FTPD_URL")

    directoryFlag := flag.String("dir", "", "specify directory (don't connect to server)")
    internalFlag := flag.Bool("int", false, "specify internal connection (don't connect to server)")
    flag.Parse()

    var saveDir = ""
    homeDir, _ := os.UserHomeDir()
    ftpdPath := filepath.Join(homeDir, ".ftpd")
    if *directoryFlag == "" {
        if _, err := os.Stat(ftpdPath); err != nil {
            saveDir = ftpdPath
        } else {
            saveDir, _ = os.Getwd()
        }
    } else {
        saveDir = *directoryFlag
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    interrupt := make(chan os.Signal, 1)
    signal.Notify(interrupt, os.Interrupt)

    u := url.URL{Scheme: ftpdScheme, Host: ftpdUrl, Path: "/connect"}

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
        ctx: ctx,
        ftpdDir: saveDir,
        ws: conn,
        tcpAddr: tcpAddr,
        internal: *internalFlag,
        filename: make(chan string, 1),
        ready: make(chan bool, 1),
    }
    cfg.ws.WriteMessage(websocket.TextMessage, []byte("TCP IP " + tcpAddr))

    go startRepl(cfg)

    done := make(chan struct{})

    go func() {
        defer close(done)
        for {
            msgType, message, err := conn.ReadMessage()
            fmt.Print("\033[G\033[K")
            if err != nil {
                log.Printf("error reading from websocket: %s\n", err)
                return
            }
            if msgType == websocket.TextMessage {
                msgArgs := strings.Fields(string(message))
                log.Printf("received: %s\n", message)
                if msgArgs[0] == "Connected" {
                    cfg.peer = msgArgs[len(msgArgs)-1]
                    fmt.Printf("    -> connected to room %s\n", cfg.peer)
                } else if msgArgs[0] == "Disconnecting" {
                    cfg.peer = ""
                    fmt.Printf("    -> disconnected from %s\n", msgArgs[len(msgArgs)-1])
                } else if msgArgs[0] == "Sending" {
                    filename := msgArgs[len(msgArgs) - 1]
                    fmt.Printf("Receive file %s. Alternative name: ", filename)
                    cfg.filename <- filename
                    cfg.ready <- false
                }
            } else {
                log.Printf("received: %x\n", message)
                if len(cfg.filename) >= 0 {
                    file, err := os.Create(<-cfg.filename)
                    if err != nil {
                        log.Printf("error writing file: %s\n", err)
                        continue
                    }
                    cfg.file = file
                }
                if len(cfg.ready) == 0 {
                    log.Print("unable to process binary data from websocket")
                    continue
                }
                writeFile(message)  //TODO: create function to write to file
            }
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
