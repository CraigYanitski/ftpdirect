package main

import (
	//"encoding/json"
	"fmt"
	"log"
	"net/http"

    "github.com/gorilla/websocket"
)

func main() {
    mux := http.NewServeMux()

    mux.HandleFunc("/connect", handleConnection)

    port := "8080"
    server := http.Server{Addr: ":"+port, Handler: mux}

    fmt.Printf("Serving on port: %s\n", port)
    log.Fatal(server.ListenAndServe())
}

type Connection struct {
    IpAddr  string  `json:"ip_addr"`
    Port    string  `json:"port"`
}

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

func handleConnection(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "unable to upgrade to websocket", err)
    }
    defer conn.Close()

    remoteAddr := conn.RemoteAddr()
    localAddr := conn.LocalAddr()

    log.Println("New websocket connection made")
    log.Printf("Local IP: %s\n", localAddr)
    log.Printf("Remote IP: %s\n", remoteAddr)

    // c := &Connection{}
    // decoder := json.NewDecoder(r.Body)
    // err = decoder.Decode(c)
    // if err != nil {
    //     respondWithError(w, http.StatusInternalServerError, "unable to unmarshal connection", err)
    // }

    // respondWithJSON(w, http.StatusOK, c)
    
    for {
        _, message, err := conn.ReadMessage()
        if err != nil {
            log.Println("Error reading message from websocket.")
            break
        }
        fmt.Printf("Received: %s\n", message)
        // err = conn.WriteMessage(websocket.TextMessage, message)
        // if err != nil {
        //     log.Println("Error echoing received message")
        //     break
        // }
    }
}

