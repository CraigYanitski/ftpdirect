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
    var arg string
    scanner := bufio.NewScanner(os.Stdin)
    time.Sleep(1*time.Second)
    for {
        if cfg.peer == "" {
            cfg.prompt = "Headless -> "
        } else {
            cfg.prompt = "\033[38;5;28m" + cfg.peer + "\033[0m -> "
        }
        fmt.Print(cfg.prompt)
        scanner.Scan()
        err := scanner.Err()
        if err != nil {
            log.Fatalf("unable to read scanner input: %s", err)
        }
        input := scanner.Text()
        if (strings.TrimSpace(input) == "") && (len(cfg.filename) == 0) {
            fmt.Print("\033[1F\033[K")
            continue
        }
        // fmt.Print("\033[1F\033[K"+cfg.prompt+input+"\n")
        commands := strings.Split(strings.TrimSpace(input), " ")
        command := strings.ToLower(commands[0])
        if _, ok := getCommands()[command]; ok {
            if len(commands) > 1 {
                arg = strings.Join(commands[1:], " ")
            } else {
                arg = ""
            }
            err = getCommands()[command].callback(cfg, arg)
            if err != nil {
                log.Println(err)
            }
        } else if len(cfg.filename) > 0 {
            if strings.TrimSpace(input) != "" {
                <-cfg.filename
                cfg.filename <- input
            }
            cfg.ready <- true
        } else if strings.Contains("yesno", input){
            var addrCmd = ""
            if cfg.peer != "" {
                addrCmd = "roomID " + cfg.peer + " "
            }
            cfg.ws.WriteMessage(websocket.TextMessage, []byte(addrCmd + input))
        }
        time.Sleep(1*time.Second)
    }
}
