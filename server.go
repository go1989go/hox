package hox

import (
	"net"
	"encoding/base64"
	"log"
	"bufio"
	"crypto/tls"
	"golang.org/x/time/rate"
	"strings"
)

type Server struct {
	listener   net.Listener
	addr       string
	credential string
	cert       string
	key        string
	maxSpeed   rate.Limit
}

func NewServer(addr, credential string, cert, key string, maxSpeed float64) *Server {
	return &Server{addr: addr, credential: base64.StdEncoding.EncodeToString([]byte(credential)), cert: cert, key: key, maxSpeed: rate.Limit(maxSpeed)}
}

func (s *Server) Start() {
	var err error
	if s.cert != "" && s.key != "" {
		pem, err := tls.LoadX509KeyPair(s.cert, s.key)
		if err != nil {
			log.Printf("tls load err :: %s", err.Error())
			return
		} else {
			config := &tls.Config{Certificates: []tls.Certificate{pem}}
			s.listener, err = tls.Listen("tcp", s.addr, config)
			log.Printf("tls on c=%s, k=%s\n", s.cert, s.key)
		}
	} else {
		s.listener, err = net.Listen("tcp", s.addr)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer s.listener.Close()
	if s.credential != "" {
		log.Printf("user %s for auth \n", s.credential)
	}
	log.Printf("TCP server listen on %s, Max speed %v\n", s.addr, s.maxSpeed)
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go s.newConn(conn).handle()
	}
}

func (s *Server) newConn(rwc net.Conn) *conn {
	return &conn{
		server:   s,
		rwc:      rwc,
		brc:      bufio.NewReader(rwc),
		maxSpeed: s.maxSpeed,
	}
}
func (s *Server) isAuth() bool {
	return s.credential != ""
}

func (s *Server) validateAuth(basicCredential string) bool {
	c := strings.Split(basicCredential, " ")
	if len(c) == 2 && strings.EqualFold(c[0], "Basic") && c[1] == s.credential {
		return true
	}
	return false
}