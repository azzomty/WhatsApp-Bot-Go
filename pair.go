package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run pair.go <phone_number>")
		return
	}
	phone := os.Args[1]

	dbLog := waLog.Stdout("Database", "ERROR", true)
	container, err := sqlstore.New(context.Background(), "sqlite3", "file:whatsapp_v3.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}

	deviceStore := container.NewDevice()
	clientLog := waLog.Stdout("Client", "INFO", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	err = client.Connect()
	if err != nil {
		fmt.Println("Connect error:", err)
		return
	}

	code, err := client.PairPhone(context.Background(), phone, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
	if err != nil {
		fmt.Println("Error pairing:", err)
		return
	}
	fmt.Printf("Pairing code for %s: %s\n", phone, code)
	
	fmt.Println("Waiting to finish login...")
	for client.Store.ID == nil {
	    // Wait
	}
	fmt.Println("Successfully logged in device:", client.Store.ID)
	fmt.Println("Keeping the connection open to finish syncing... DO NOT KILL YET.")
	
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
