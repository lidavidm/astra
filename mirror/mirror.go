package main

import (
	"context"
	"log"
	"net/http"

	"flag"

	logclient "github.com/google/certificate-transparency/go/client"
	"github.com/google/certificate-transparency/go/jsonclient"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	var source string
	var dest string
	flag.StringVar(&source, "source", "localhost:6962", "The URL to the source CT server.")
	flag.StringVar(&dest, "dest", "localhost:6962", "The URL to the destination CT server.")

	flag.Parse()

	client := http.DefaultClient
	options := jsonclient.Options{}
	ctx := context.Background()

	log.Print("Creating clients...")
	srcclient, err := logclient.New(source, client, options)
	if err != nil {
		log.Fatal(err)
	}

	destclient, err := logclient.New(dest, client, options)
	if err != nil {
		log.Fatal(err)
	}

	sth, err := srcclient.GetSTH(ctx)
	if err != nil {
		log.Fatal(err)
	}

	numEntries := int64(sth.TreeSize)
	// Limit # of entries for testing
	if numEntries > 10 {
		numEntries = 10
	}

	certs, err := srcclient.GetEntries(ctx, 0, numEntries)
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range certs {
		timestamp, err := destclient.AddChain(ctx, entry.Chain)
		if err != nil {
			log.Print("Error adding certificate:", err)
		} else {
			log.Print("Got timestamp:", timestamp)
		}
	}
}
