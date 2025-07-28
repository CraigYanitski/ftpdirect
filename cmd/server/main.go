package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

type apiConfig struct {
    connections  map[string][]*websocket.Conn
    mu           sync.Mutex
}

func main() {
    godotenv.Load()
    port := os.Getenv("FTPD_PORT")

    cfg := apiConfig{
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
        // process text message
        if msgType == websocket.TextMessage {
            // log message
            fmt.Printf("Received: %s\n", message)

            // check if connection command
            connectArgs := strings.Fields(strings.TrimSpace(string(message)))
            var rID = ""
            if (connectArgs[0] == "Connect") && (len(connectArgs) <= 2) {
                if len(connectArgs) == 2 {
                    rID = connectArgs[1]
                } else {
                    rID = ""
                }
                roomID := cfg.connect(
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
            } else if (connectArgs[0] == "Disconnect") && (len(connectArgs) == 2) {
                cfg.disconnect(connectArgs[1])
            } else if ((connectArgs[0] == "TCP") || (connectArgs[0] == "IP")) && (len(connectArgs) <= 3) {
                continue
            } else {
                // all other communication should begin with room ID
                conns, ok := cfg.connections[connectArgs[0]]
                if !ok {
                    log.Printf("Requested room ID %s not found.\n", connectArgs[0])
                    conn.WriteMessage(
                        websocket.TextMessage, 
                        []byte(fmt.Sprintf("Invalid room ID %s", connectArgs[0])),
                    )
                }
                for _, c := range conns {
                    if c != conn {
                        c.WriteMessage(websocket.TextMessage, message)
                    }
                }
            }
        } else {
            // otherwise send binary data to peer
            for _, c := range cfg.connections[roomID] {
                if c != conn {
                    c.WriteMessage(websocket.BinaryMessage, message)
                }
            }
        }
    }
}

func (cfg *apiConfig) disconnect(roomID string) {
    /* function to disconnect a client from a room; currently deletes room */
    cfg.mu.Lock()
    defer cfg.mu.Unlock()
    for _, c := range cfg.connections[roomID] {
        c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("disconnecting from room %s", roomID)))
    }
    delete(cfg.connections, roomID)
    return
}

func (cfg *apiConfig) connect(conn *websocket.Conn, roomID string) string {
    /* function to connect clients to peers */
    cfg.mu.Lock()
    defer cfg.mu.Unlock()
    if room, ok := cfg.connections[roomID]; len(roomID) > 0 && ok {
        var resp = ""
        for {
            room[0].WriteMessage(
                websocket.TextMessage, 
                []byte(fmt.Sprintf("%s attempting to join (y|n)", conn.RemoteAddr())),
            )
            _, msg, err := room[0].ReadMessage()
            if err != nil {
                log.Printf("Error verifying connection to peer to peer room: %s", err)
                return ""
            }
            resp = strings.TrimSpace(strings.ToLower(string(msg)))
            if (resp == "yes") || (resp == "n") {
                break
            } else if (resp == "no") || (resp == "n") {
                return "rejected"
            }
        }
        room = append(room, conn)
        cfg.connections[roomID] = room
    } else {
        roomID = uuid.New().String()
        cfg.connections[roomID] = []*websocket.Conn{conn}
    }
    return roomID
}

