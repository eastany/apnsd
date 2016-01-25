package main

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/eastany/apnsd/common/feedback"
)

const (
	PACKET_MAX_LEN = 4096
	headerLen      = 5
	testToken      = "e93b7686988b4b5fd334298e60e73d90035f6d12628a80b4029bde0dec514df9"
)

var all_cnt uint64

type TlsTCPServer interface {
	Start() error
	RunLoop()
}

type FKServer struct {
	TlsCfg    *tls.Config
	servePort int
	l         net.Listener
	acptNum   int
	wa        *sync.WaitGroup
	handle    func(net.Conn)
}

func createTlsConfig(crt, key string) *tls.Config {
	ca_b, err := ioutil.ReadFile("ca.pem")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	ca, err := x509.ParseCertificate(ca_b)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	priv_b, err := ioutil.ReadFile("ca.key")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	priv, err := x509.ParsePKCS1PrivateKey(priv_b)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	pool := x509.NewCertPool()
	pool.AddCert(ca)
	cert := tls.Certificate{
		Certificate: [][]byte{ca_b},
		PrivateKey:  priv,
	}
	config := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
		ClientCAs:    pool,
		ServerName:   "eastany.com",
	}
	config.Rand = rand.Reader
	return config
}

func NewServerWithTls(port int, cfg *tls.Config, handle func(net.Conn)) *FKServer {
	return &FKServer{
		servePort: port,
		acptNum:   4,
		wa:        new(sync.WaitGroup),
		TlsCfg:    cfg,
		handle:    handle,
	}
}

func (fs *FKServer) Start() (err error) {
	fs.l, err = tls.Listen("tcp4", fmt.Sprintf(":%d", fs.servePort), fs.TlsCfg)
	return
}

func (fs *FKServer) RunLoop() {
	for i := 0; i < fs.acptNum; i++ {
		var index = i
		go fs.serve(index)
	}
	select {}
}

func (fs *FKServer) serve(id int) {
	if id == 3 {
		return
	}
	id += 1
	fmt.Printf("serve %d\n", id)
	fs.wa.Add(1)
	defer fs.wa.Done()

	for {
		c, err := fs.l.Accept()
		if err != nil {
			continue
		}
		go fs.handle(c)
	}
}

func handleNotification(c net.Conn) {
	defer c.Close()
	tlsConn, ok := c.(*tls.Conn)
	if ok {
		if err := tlsConn.Handshake(); err != nil {
			fmt.Printf("handshake err : %v\n", err)
			return
		}
		state := tlsConn.ConnectionState()
		fmt.Println(state.ServerName)
		fmt.Println(state.HandshakeComplete)
	}
	for {
		buffer := make([]byte, PACKET_MAX_LEN)
		n, err := io.ReadFull(tlsConn, buffer[:headerLen])
		if err != nil {
			fmt.Println(err)
			return
		}
		if n != headerLen {
			fmt.Println("Bad header")
			continue
		}
		pktLen := binary.BigEndian.Uint32(buffer[1:headerLen])
		n, err = io.ReadFull(tlsConn, buffer[headerLen:pktLen+headerLen])
		if err != nil {
			fmt.Println(err)
			return
		}
		if uint32(n) != pktLen {
			fmt.Printf("Bad Data for %d \n", n)
			continue
		}
		atomic.AddUint64(&all_cnt, 1)
		if atomic.LoadUint64(&all_cnt)%10000 == 0 {
			fmt.Printf("%d -- %d\n", time.Now().Unix(), atomic.LoadUint64(&all_cnt))
		}
	}
}

func handleFeedback(c net.Conn) {
	defer c.Close()

	for {
		fd := feedback.NewFeedback(testToken)
		data, err := fd.ToBytes()
		if err == nil {
			c.Write(data)
		}
		time.Sleep(time.Second * 1)
	}
}

func FactoryTlsTcpServer() (s []TlsTCPServer) {
	cfg := createTlsConfig("ca.pem", "ca.key")
	if cfg == nil {
		return
	}
	ntServer := NewServerWithTls(9001, cfg, handleNotification)
	fdServer := NewServerWithTls(9002, cfg, handleFeedback)
	s = append(s, ntServer)
	s = append(s, fdServer)
	return
}

func main() {
	ss := FactoryTlsTcpServer()
	for _, s := range ss {
		err := s.Start()
		if err != nil {
			fmt.Printf("%v\n", err)
			continue
		}
		go s.RunLoop()
	}
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Kill, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT)
	<-sigCh
}
