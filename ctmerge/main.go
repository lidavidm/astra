package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	ct "github.com/google/certificate-transparency/go"
	"github.com/google/certificate-transparency/go/client"
	"github.com/google/certificate-transparency/go/jsonclient"
	"golang.org/x/net/context"
)

func createClient(url string) (*client.LogClient, error) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}
	httpCli := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSHandshakeTimeout:   30 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			MaxIdleConnsPerHost:   10,
			DisableKeepAlives:     false,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	return client.New(url, httpCli, jsonclient.Options{})
}

func main() {
	// Begin Processing Flags
	startIdxFlag := flag.Int("start", 0, "The 0-based leaf entry index to start at")
	blkSizeFlag := flag.Int("size", 64, "Amount of entries to query each iteration")
	flag.Usage = func() {
		fmt.Printf("Usage: %s [options...] <source CT server base URL> <dest CT server base URL>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}
	cturlSrc, cturlDest := flag.Arg(0), flag.Arg(1)
	startIdx := *startIdxFlag
	blkSize := int64(*blkSizeFlag)
	// Flag Processing Done

	// Use Google's CT Go libraries to access the two servers
	logSrc, err := createClient(cturlSrc)
	if err != nil {
		log.Fatalf("CT Log client for source url error: %v", err)
	}
	logDest, err := createClient(cturlDest)
	if err != nil {
		log.Fatalf("CT Log client for dest url error: %v", err)
	}

	// Find out how many leaves exist at the source CT log
	sth, err := logSrc.GetSTH(context.TODO())
	if err != nil {
		log.Fatalf("unable to retrieve Log Client STH for source URL: %v", err)
	}
	lenLogSrc := int64(sth.TreeSize)
	if int64(startIdx) >= lenLogSrc {
		return
	}

	log.Printf("Attempting to send %v certificates to the destination CT server.\n", lenLogSrc)
	for i := int64(startIdx); i < lenLogSrc; i += blkSize {
		// upper = MAX ( last_index_in_block , tree_size - 1 )
		upper := int64(i + (blkSize - 1))
		if upper >= lenLogSrc {
			upper = lenLogSrc - 1
		}
		entries, err := logSrc.GetEntries(context.TODO(), i, upper)
		if err != nil {
			log.Printf("unable to call get entries [%v,%v]: %v", i, upper, err)
			continue
		}
		gotEntireBlock := false // Whether or not we got every cert that we requested.
		for j, entry := range entries {
			var err error
			var sct *ct.SignedCertificateTimestamp
			switch entry.Leaf.TimestampedEntry.EntryType {
			case ct.X509LogEntryType:
				chain := []ct.ASN1Cert{*entry.Leaf.TimestampedEntry.X509Entry}
				chain = append(chain, entry.Chain...)
				sct, err = logDest.AddChain(context.TODO(), chain)
			case ct.PrecertLogEntryType:
				chain := entry.Chain
				sct, err = logDest.AddPreChain(context.TODO(), chain)
			}
			if err != nil {
				log.Printf("unable to send certificate #%v: %v", i+int64(j), err)
				continue
			}
			if j == (int(blkSize) - 1) { // last Entry of Block
				log.Printf("sent certificate #%v - got TS %v\n", i+int64(j), sct.Timestamp)
				gotEntireBlock = true
			}
		}
		if !gotEntireBlock {
			log.Printf("warning: did not receive all entries requested. use smaller -size arg")
		}
	}
}
