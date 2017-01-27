package main

import (
	"crypto/tls"
	"github.com/majestrate/tls2socks/server"
	"log"
	"net"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		log.Printf("usage: %s cert.pem key.pem", os.Args[0])
		return
	}

	cert, err := tls.LoadX509KeyPair(os.Args[1], os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	sock, err := net.Listen("tcp", ":6697")
	if err != nil {
		log.Fatal(err)
	}
	defer sock.Close()
	s := &server.Server{
		TLSConf: &tls.Config{
			GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
				return &cert, nil
			},
		},
		Sock:     sock,
		Upstream: "127.0.0.1:4447",
		Port:     6667,
	}
	err = s.Run()
}
