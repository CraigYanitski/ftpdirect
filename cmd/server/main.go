package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

func handleConnection(w http.ResponseWriter, r *http.Request) {
    c := &Connection{}
    decoder := json.NewDecoder(r.Body)
    err := decoder.Decode(c)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "unable to unmarshal connection", err)
    }

    respondWithJSON(w, http.StatusOK, c)
}
