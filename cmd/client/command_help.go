package main

import "fmt"

func commandHelp(cfg *apiConfig, arg string) error {
    fmt.Print("\nWelcome to FTP direct!")
    fmt.Print("\n----------------------\nUsage\n\n")
    for _, command := range getCommands() {
        fmt.Printf("%s: %s\n", command.name, command.description)
    }
    fmt.Print("\n")
    return nil
}
