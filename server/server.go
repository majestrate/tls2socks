package server

import (
	"crypto/tls"
	"encoding/binary"
	"io"
	"log"
	"net"
	"strings"
)

type Server struct {
	TLSConf  *tls.Config
	Sock     net.Listener
	Upstream string
	Port     uint16
}

func (s *Server) runConn(conn *tls.Conn) {
	var c net.Conn
	err := conn.Handshake()
	if err == nil {
		name := conn.ConnectionState().ServerName
		parts := strings.Split(name, ".")
		if len(parts) > 2 {
			host := ""
			for idx, part := range parts {
				if idx == len(parts)-2 {
					host += part
					break
				} else {
					host += part + "."
				}
			}
			log.Println("host requested", host)
			// valid
			c, err = net.Dial("tcp", s.Upstream)
			if err == nil {
				defer c.Close()
				// do socks 4a handshake
				ident := []byte("socks")
				req := make([]byte, 8+len(ident)+1+len(host)+1)
				req[0] = 0x04
				req[1] = 0x01
				binary.BigEndian.PutUint16(req[2:], s.Port) // irc port
				req[7] = 0x01
				copy(req[8:], ident)
				copy(req[8+1+len(ident):], []byte(host))
				_, err = c.Write(req)
				if err == nil {
					var resp [8]byte
					_, err = io.ReadFull(c, resp[:])
					if err == nil {
						// handshake completed
						if resp[1] == 0x5a {
							// good
							chnl := make(chan error)

							// start forwarding

							go func() {
								var buf [1024]byte
								_, e := io.CopyBuffer(conn, c, buf[:])
								chnl <- e
							}()
							go func() {
								var buf [1024]byte
								_, e := io.CopyBuffer(c, conn, buf[:])
								chnl <- e
							}()
							err1 := <-chnl
							err2 := <-chnl
							close(chnl)
							if err1 == nil {
								err = err2
							} else {
								err = err1
							}
						}
					}
				}
			}
		}
	}
	conn.Close()
}

func (s *Server) Run() (err error) {
	var c net.Conn
	for err == nil {
		c, err = s.Sock.Accept()
		if err == nil {
			tc := tls.Server(c, s.TLSConf)
			go s.runConn(tc)
		}
	}
	return
}
