package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalln("Usage:", os.Args[0], "<CT server>")
	}
	cturl := os.Args[1]

	if !strings.HasPrefix(cturl, "http://") && !strings.HasPrefix(cturl, "https://") {
		cturl = "http://" + cturl
	}

	rootsurl := cturl
	if !strings.HasSuffix(rootsurl, "ct/v1/get-roots") {
		// Take base CT server and append ./ct/v1/get-roots path.
		if !strings.HasSuffix(cturl, "/") {
			cturl += "/"
		}
		rel, _ := url.Parse("ct/v1/get-roots")
		base, _ := url.Parse(cturl)
		rootsurl = base.ResolveReference(rel).String()
	}
	log.Println("Reading from ", rootsurl)

	resp, err := http.Get(rootsurl)
	if err != nil {
		log.Fatal(err)
	}
	respJSON, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	buf, _ := convertToPEM(respJSON)
	fname := "roots.pem"
	f, err := os.Create(fname)
	if err != nil {
		log.Fatal(err)
	}
	buf.WriteTo(f)
	f.Close()
	log.Println("Wrote out to", fname)
}

func convertToPEM(jsonblob []byte) (*bytes.Buffer, error) {
	type CertificateList struct {
		Certificates []string
	}
	var clist CertificateList
	err := json.Unmarshal(jsonblob, &clist)
	if err != nil {
		return nil, err
	}
	log.Printf("Found %d root certificates to be saved\n", len(clist.Certificates))
	// Encode into PEM format and word-wrap 64-chars per line.
	prefix := []byte("-----BEGIN CERTIFICATE-----\n")
	suffix := []byte("-----END CERTIFICATE-----")
	buf := new(bytes.Buffer)
	for idx, cert := range clist.Certificates {
		buf.Write(prefix)
		for i := 0; i < len(cert)/64; i++ {
			buf.WriteString(cert[i*64 : (i+1)*64])
			buf.WriteRune('\n')
		}
		lastBlock := cert[(len(cert)/64)*64:]
		if lastBlock != "" {
			buf.WriteString(lastBlock)
			buf.WriteRune('\n')
		}
		buf.Write(suffix)
		if idx != len(clist.Certificates)-1 {
			buf.WriteRune('\n')
		}
	}

	return buf, nil
}
