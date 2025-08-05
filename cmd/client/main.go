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
    prompt     string
    peerConn   *net.TCPConn
    file       *os.File
    internal   bool
    filename   chan string
    ready      chan bool
}

func main() {
    godotenv.Load()
    ftpdScheme := os.Getenv("FTPD_SCHEME")
    ftpdUrl := os.Getenv("FTPD_URL")
    ftpdEndpoint := os.Getenv("FTPD_ENDPOINT")

    directoryFlag := flag.String("dir", "", "specify directory (don't connect to server)")
    internalFlag := flag.Bool("int", false, "specify internal connection (don't connect to server)")
    flag.Parse()

    var saveDir = ""
    homeDir, _ := os.UserHomeDir()
    ftpdPath := filepath.Join(homeDir, ".ftpd")
    if *directoryFlag == "" {
        if _, err := os.Stat(ftpdPath); os.IsNotExist(err) {
            saveDir, _ = os.Getwd()
        } else {
            saveDir = ftpdPath
        }
    } else {
        saveDir = *directoryFlag
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    interrupt := make(chan os.Signal, 1)
    signal.Notify(interrupt, os.Interrupt)

    u := url.URL{Scheme: ftpdScheme, Host: ftpdUrl, Path: ftpdEndpoint}

    conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
    if err != nil {
        log.Fatalln(err)
    }
    defer conn.Close()

    cfg := &apiConfig{
        ctx: ctx,
        ftpdDir: saveDir,
        ws: conn,
        internal: *internalFlag,
        filename: make(chan string, 1),
        ready: make(chan bool, 1),
    }

    err = cfg.startTCPServer()
    if err != nil {
        log.Println("error starting TCP server")
        return
    }
    log.Printf("TCP listening on %s\n", cfg.tcpAddr)
    cfg.ws.WriteMessage(
        websocket.TextMessage, 
        []byte("TCP IP " + cfg.tcpAddr),
    )

    go startRepl(cfg)
    go wsListener(cfg)

    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-cfg.ctx.Done():
            return
        case <-interrupt:
            commandExit(cfg, "")
            return
        }
    }
}

func (cfg *apiConfig) startTCPServer() error {
    listener, err := net.Listen("tcp", ":0")
    if err != nil {
        return err
    }
    port := listener.Addr().(*net.TCPAddr).Port
    go func() {
        for {
            conn, err := listener.Accept()
            if err != nil{
                continue
            }
            err = cfg.createFile(time.Now().Format("2006-01-02_15:04:05"))
            buf := make([]byte, 1<<20)
            for {
                n, err := conn.Read(buf)
                if err != nil {
                    break
                }
                cfg.writeFile(buf[:n])
            }
            cfg.file.Close()
            cfg.file = nil
        }
    }()
    cfg.tcpAddr = fmt.Sprintf("%s:%d", getLocalIP(), port)
    return nil
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

