package main

import (
    "fmt"
)

func commandExit(cfg *apiConfig, arg string) error {
    fmt.Println("Closing FTP direct... Goodbye!")
    cfg.ctx.Done()
    return nil
}
