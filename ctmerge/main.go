package main

import (
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
	if len(os.Args) != 3 {
		log.Fatalln("Usage:", os.Args[0], "<source CT server base URL> <dest CT server base URL>")
	}
	cturlSrc, cturlDest := os.Args[1], os.Args[2]

	logSrc, err := createClient(cturlSrc)
	if err != nil {
		log.Fatalf("CT Log client for source url error: %v", err)
	}
	logDest, err := createClient(cturlDest)
	if err != nil {
		log.Fatalf("CT Log client for dest url error: %v", err)
	}

	sth, err := logSrc.GetSTH(context.TODO())
	if err != nil {
		log.Fatalf("unable to retrieve Log Client STH for source URL: %v", err)
	}
	lenLogSrc := int64(sth.TreeSize)
	log.Printf("Attempting to send %v certificates to the destination CT server.\n", lenLogSrc)
	var i int64
	for i = 0; i < lenLogSrc; i += 8 {
		upper := int64(i + 7)
		if upper > lenLogSrc {
			upper = lenLogSrc
		}
		entries, err := logSrc.GetEntries(context.TODO(), i, upper)
		if err != nil {
			log.Printf("unable to call get entries [%v,%v]: %v", i, upper, err)
			continue
		}
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
			log.Printf("sent certificate #%v to get: %v\n", i+int64(j), sct.String())
		}
	}
}
