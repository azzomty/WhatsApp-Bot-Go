package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dbLog := waLog.Stdout("Database", "ERROR", true)
	container, err := sqlstore.New(context.Background(), "sqlite3", "file:whatsapp_v3.db?_foreign_keys=on&_busy_timeout=60000&_journal_mode=WAL", dbLog)
	if err != nil {
		panic(err)
	}

	deviceStore := container.NewDevice()
	clientLog := waLog.Stdout("Client", "INFO", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	
	err = client.Connect()
	if err != nil {
		panic(err)
	}

	phone := os.Args[1]
	code, err := client.PairPhone(context.Background(), phone, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n\nPairing code for %s: %s\n\n", phone, code)
	
	for client.Store.ID == nil {
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Println("Connected! Syncing history for 1 min...")
	time.Sleep(1 * time.Minute)
	fmt.Println("Done!")
}
