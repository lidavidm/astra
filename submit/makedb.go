package main

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"io/ioutil"
	"log"
	"math/big"

	"flag"

	"github.com/cloudflare/cfssl/certdb"
	"github.com/cloudflare/cfssl/certdb/dbconf"
	"github.com/cloudflare/cfssl/certdb/sql"
	"github.com/cloudflare/cfssl/helpers"

	"time"

	"encoding/pem"

	_ "github.com/mattn/go-sqlite3"
)

func makeCertificate(root *x509.Certificate, rootKey crypto.Signer) (*big.Int, []byte, []byte, error) {
	serialNumberRange := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberRange)
	if err != nil {
		return nil, nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Cornell CS 5152"},
		},
		AuthorityKeyId: []byte{42, 42, 42, 42},
	}
	cert, err := x509.CreateCertificate(rand.Reader, &template, root, rootKey.Public(), rootKey)
	return serialNumber, []byte{42, 42, 42, 42}, cert, err
}

func main() {
	var dbConfig string
	flag.StringVar(&dbConfig, "dbConfig", "", "The certdb config path.")

	flag.Parse()

	db, err := dbconf.DBFromConfig(dbConfig)
	if err != nil {
		log.Print(dbConfig)
		log.Fatal("Could not load certdb: ", err, dbConfig)
	}

	dbAccessor := sql.NewAccessor(db)

	rootData, err := ioutil.ReadFile("fake-ca.cert")
	if err != nil {
		log.Fatal(err)
	}
	root, err := helpers.ParseCertificatePEM(rootData)
	if err != nil {
		log.Fatal(err)
	}

	rootKeyData, err := ioutil.ReadFile("fake-ca.privkey.pem")
	if err != nil {
		log.Fatal(err)
	}
	rootKey, err := helpers.ParsePrivateKeyPEMWithPassword(rootKeyData, []byte("gently"))
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 20; i++ {
		serial, aki, cert, err := makeCertificate(root, rootKey)

		if err != nil {
			log.Fatal(err)
		}

		if err := dbAccessor.InsertCertificate(certdb.CertificateRecord{
			Serial: serial.Text(16),
			AKI:    hex.EncodeToString(aki),
			PEM: string(pem.EncodeToMemory(&pem.Block{
				Type:  "CERTIFICATE",
				Bytes: cert,
			})),
			Expiry: time.Now().Add(365 * 24 * time.Hour),
		}); err != nil {
			log.Println(err)
			continue
		}
	}
}
