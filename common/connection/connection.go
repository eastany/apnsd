package connection

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"strings"
	"sync/atomic"
)

type TlsConnection struct {
	conn      *tls.Conn
	rootCA    *x509.Certificate
	crt       tls.Certificate
	host      string
	CloseFlag uint32
}

func NewTlsConnWithCert(root *x509.Certificate,
	cert tls.Certificate, host string) *TlsConnection {
	return &TlsConnection{
		rootCA:    root,
		crt:       cert,
		host:      host,
		CloseFlag: 1,
	}
}

func (c *TlsConnection) Connect() {
	// TODO lock
	if atomic.LoadUint32(&c.CloseFlag) == 0 {
		return
	}
	hs := strings.Split(c.host, ":")
	peerDomain := hs[0]

	pool := x509.NewCertPool()
	if c.rootCA != nil {
		pool.AddCert(c.rootCA)
	}
	config := &tls.Config{
		RootCAs:            pool,
		Certificates:       []tls.Certificate{c.crt},
		InsecureSkipVerify: false,
		ServerName:         peerDomain,
	}

	conn, err := net.Dial("tcp", c.host)
	if err != nil {
		fmt.Println(err)
		return
	}
	tlsConn := tls.Client(conn, config)
	if err = tlsConn.Handshake(); err != nil {
		fmt.Println(err)
		return
	}
	c.conn = tlsConn
	atomic.StoreUint32(&c.CloseFlag, 0)
	return
}

func (c *TlsConnection) ReadFull(buf []byte) (int, error) {
	return io.ReadFull(c.conn, buf)
}

func (c *TlsConnection) WriteAll(buf []byte) (int, error) {
	return c.conn.Write(buf)
}

func (c *TlsConnection) Close() {
	c.conn.Close()
	atomic.StoreUint32(&c.CloseFlag, 1)
}
