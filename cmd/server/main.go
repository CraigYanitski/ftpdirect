package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	// "time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

const (
    confirmOption = "yes"
    declineOption = "no"
    decisionPrompt = "(" + confirmOption + "|" + declineOption + ")"
)

type apiConfig struct {
    connections  map[string][]*websocket.Conn
    signals      map[string]chan bool
    mu           sync.Mutex
}

func main() {
    godotenv.Load()
    port := os.Getenv("FTPD_PORT")

    cfg := apiConfig{
        connections: make(map[string][]*websocket.Conn),
        signals: make(map[string]chan bool),
        mu: sync.Mutex{},
    }

    mux := http.NewServeMux()
    mux.HandleFunc("/connect", cfg.handleConnection)

    server := http.Server{Addr: ":"+port, Handler: mux}

    fmt.Printf("Serving on port: %s\n", port)
    log.Fatal(server.ListenAndServe())
}

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

func (cfg *apiConfig) handleConnection(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "unable to upgrade to websocket", err)
    }
    defer func(){
        log.Printf("%s disconnected", conn.RemoteAddr())
        conn.Close()
    }()



    remoteAddr := conn.RemoteAddr()
    localAddr := conn.LocalAddr()

    log.Println("New websocket connection made")
    log.Printf("Local IP: %s\n", localAddr)
    log.Printf("Remote IP: %s\n", remoteAddr)

    conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("public IP %s\n", remoteAddr)))

    var roomID string
    
    // loop forever to check for server commands, otherwise forward the messages
    for {
        msgType, message, err := conn.ReadMessage()
        if err != nil {
            log.Printf("Error reading message from websocket: %s\n", err)
            break
        }
        switch msgType {
        // process text message
        case websocket.TextMessage:
            // log message
            fmt.Printf("Received: %s\n", message)

            // check if connection command
            msgArgs := strings.Fields(strings.TrimSpace(string(message)))
            var rID = ""
            if strings.Contains(decisionPrompt, strings.ToLower(msgArgs[0])) && (len(msgArgs) == 1) {
                conn.WriteMessage(websocket.TextMessage, message)
            } else if (msgArgs[0] == "Connect") && (len(msgArgs) <= 2) {
                if len(msgArgs) == 2 {
                    rID = msgArgs[1]
                } else {
                    rID = ""
                }
                roomID = cfg.connect(
                    conn, 
                    rID,
                )
                
                // Check for error (blank) or rejection, else success
                if roomID == "" {
                    conn.WriteMessage(
                        websocket.TextMessage,
                        []byte(fmt.Sprintf("Unable to join room %s", rID)),
                    )
                } else if roomID == "rejected" {
                    conn.WriteMessage(
                        websocket.TextMessage,
                        []byte(fmt.Sprintf("Rejected from room %s", rID)),
                    )
                    roomID = ""
                } else {
                    conn.WriteMessage(
                        websocket.TextMessage, 
                        []byte(fmt.Sprintf("Connected to room %s", roomID)),
                    )
                }
            } else if (msgArgs[0] == "Disconnect") && (len(msgArgs) == 2) {
                cfg.disconnect(msgArgs[1])
                roomID = ""
            } else if ((msgArgs[0] == "TCP") || (msgArgs[0] == "IP")) && (len(msgArgs) <= 3) {
                continue
            } else if (msgArgs[0] == "Close") && (len(msgArgs) == 0) {
                if roomID != "" {
                    cfg.disconnect(roomID)
                }
                conn.WriteMessage(websocket.CloseMessage, []byte{})
                time.Sleep(5*time.Second)
                return
            } else if msgArgs[0] == "roomID" {
                log.Printf("sending to roomID: %s\n", roomID)
                // all other communication should begin with room ID
                conns, ok := cfg.connections[msgArgs[1]]
                if !ok {
                    log.Printf("Requested room ID %s not found.\n", msgArgs[0])
                    conn.WriteMessage(
                        websocket.TextMessage, 
                        []byte(fmt.Sprintf("Invalid room ID %s", msgArgs[0])),
                    )
                }
                if (len(msgArgs) == 3) && (strings.Contains(decisionPrompt, strings.ToLower(msgArgs[2]))) {
                    if _, ok := cfg.signals[msgArgs[1]]; ok {
                        fmt.Println(roomID)
                        resp := strings.ToLower(msgArgs[2])
                        cfg.mu.Lock()
                        if (resp == confirmOption) || (resp == confirmOption[:2]) {
                            cfg.signals[msgArgs[1]] <- true
                        } else if (resp == declineOption) || (resp == declineOption[:2]) {
                            cfg.signals[msgArgs[1]] <- false
                        }
                        cfg.mu.Unlock()
                    }
                }
                // send only message (not roomID) to peers
                for _, c := range conns {
                    if c != conn {
                        c.WriteMessage(websocket.TextMessage, []byte(strings.Join(msgArgs[2:], " ")))
                    }
                }
            }
        case websocket.BinaryMessage:
            // otherwise send binary data to peer
            for _, c := range cfg.connections[roomID] {
                if c != conn {
                    c.WriteMessage(websocket.BinaryMessage, message)
                }
            }
        case websocket.CloseMessage:
            if roomID != "" {
                cfg.disconnect(roomID)
            }
            conn.WriteMessage(websocket.CloseMessage, []byte{})
            time.Sleep(5*time.Second)
            return
        }
    }
}

func (cfg *apiConfig) disconnect(roomID string) {
    /* function to disconnect a client from a room; currently deletes room */
    cfg.mu.Lock()
    defer cfg.mu.Unlock()
    for _, c := range cfg.connections[roomID] {
        c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Disconnecting from room %s", roomID)))
    }
    delete(cfg.connections, roomID)
    return
}

func (cfg *apiConfig) connect(conn *websocket.Conn, roomID string) string {
    /* function to connect clients to peers */
    if room, ok := cfg.connections[roomID]; len(roomID) > 0 && ok {
        log.Printf("Attempting to join room %s pending approval\n", roomID)
        room = append(room, conn)
        cfg.mu.Lock()
        cfg.connections[roomID] = room
        cfg.mu.Unlock()
        var resp = false
        // This loop waiting for a response can take a while, so only lock the mutex when updating connections
        for {
            cfg.mu.Lock()
            cfg.signals[roomID] = make(chan bool)
            cfg.mu.Unlock()
            room[0].WriteMessage(
                websocket.TextMessage, 
                []byte(fmt.Sprintf("%s attempting to join " + decisionPrompt, conn.RemoteAddr())),
            )
            // time.Sleep(5*time.Second)
            log.Println("waiting for response")
            // _, msg, err := conn.ReadMessage()
            // if err != nil {
            //     log.Printf("Error verifying connection to peer to peer room: %s", err)
            //     return ""
            // }
            resp = <-cfg.signals[roomID]  // = strings.TrimSpace(strings.ToLower(string(msg)))
            cfg.mu.Lock()
            delete(cfg.signals, roomID)
            cfg.mu.Unlock()
            log.Println(resp)
            if resp { //(resp == confirmOption) || (resp == confirmOption[:2]) {
                break
            } else { // if (resp == declineOption) || (resp == declineOption[:2]) {
                r := []*websocket.Conn{}
                for _, c := range room {
                    if c != conn {
                        r = append(r, c)
                    }
                }
                cfg.mu.Lock()
                cfg.connections[roomID] = r
                cfg.mu.Unlock()
                return "rejected"
            }
        }
        log.Printf("Accepted %s to room %s\n", conn.RemoteAddr(), roomID)
    } else {
        cfg.mu.Lock()
        defer cfg.mu.Unlock()
        roomID = uuid.New().String()
        cfg.connections[roomID] = []*websocket.Conn{conn}
    }
    return roomID
}

