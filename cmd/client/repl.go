package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type cliCommand struct {
    name         string
    description  string
    callback     func(cfg *apiConfig, arg string) error
}

func getCommands() map[string]cliCommand {
    return map[string]cliCommand{
        "exit": {
            name: "exit",
            description: "exit the program",
            callback: commandExit,
        },
        "help": {
            name: "help",
            description: "print available commands and descriptions",
            callback: commandHelp,
        },
        "send": {
            name: "send",
            description: "send file to peer",
            callback: commandSend,
        },
        "connect": {
            name: "connect",
            description: "initialise or connect to a peer to peer room (remote) or TCP listener (local)",
            callback: commandConnect,
        },
        "disconnect": {
            name: "disconnect",
            description: "disconnect from peer to peer room",
            callback: commandDisconnect,
        },
    }
}

func startRepl(cfg *apiConfig) {
    var prompt string
    var arg string
    scanner := bufio.NewScanner(os.Stdin)
    time.Sleep(2*time.Second)
    for {
        if cfg.peer == "" {
            prompt = "Headless -> "
        } else {
            prompt = cfg.peer + " -> "
        }
        fmt.Print(prompt)
        scanner.Scan()
        err := scanner.Err()
        if err != nil {
            log.Fatalf("unable to read scanner input: %s", err)
        }
        input := scanner.Text()
        fmt.Print("\033[1F\033[K"+prompt+input+"\n")
        commands := strings.Split(strings.ToLower(strings.TrimSpace(input)), " ")
        command := commands[0]
        if _, ok := getCommands()[command]; ok {
            if len(commands) > 1 {
                arg = commands[1]
            } else {
                arg = ""
            }
            err = getCommands()[command].callback(cfg, arg)
            if err != nil {
                log.Println(err)
            }
            continue
        } else {
            var addrCmd = ""
            if cfg.peer != "" {
                addrCmd = "roomID " + cfg.peer + " "
            }
            cfg.ws.WriteMessage(websocket.TextMessage, []byte(addrCmd + input))
            continue
        }
    }
}
