package main

import (
	"errors"
	"log"
	"os"
)

func commandSend(cfg *apiConfig, arg string) error {
    if cfg.internal {
        if cfg.peerConn == nil {
            err := errors.New("Must first connect to TCP IP address before sending data")
            log.Println(err)
            return err
        }
        _, err := os.Open(arg)
        if err != nil {
            return err
        }
        //(*cfg.peerConn).Write(file) //TODO: use function for chunked messages
    } else {
    }
    return nil
}
