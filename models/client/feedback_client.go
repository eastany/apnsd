package client

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	tlsconn "github.com/eastany/apnsd/common/connection"
	"github.com/eastany/apnsd/common/queue"
)

const (
	sendFailedQueue = "Feedback_q"
)

type FeedbackClient struct {
	Conn *tlsconn.TlsConnection
	host string
	q    queue.Queue
}

func NewFeedbackClient(host string) (fc *FeedbackClient) {
	root, err := ioutil.ReadFile("ca.pem")
	if err != nil {
		return
	}
	cert2_b, err := ioutil.ReadFile("cert2.pem")
	if err != nil {
		return
	}
	priv2_b, err := ioutil.ReadFile("cert2.key")
	if err != nil {
		return
	}
	priv2, err := x509.ParsePKCS1PrivateKey(priv2_b)
	if err != nil {
		return
	}
	rootCA, err := x509.ParseCertificate(root)
	if err != nil {
		return
	}
	cert := tls.Certificate{
		Certificate: [][]byte{cert2_b},
		PrivateKey:  priv2,
	}
	conn := tlsconn.NewTlsConnWithCert(rootCA, cert, host)
	fc = &FeedbackClient{Conn: conn}
	return
}

func (c *FeedbackClient) ReadLoop() {
	c.Conn.Connect()
	defer c.Conn.Close()

	for {
		buf := make([]byte, 38)
		_, err := c.Conn.ReadFull(buf)
		if err != nil {
			break
		}
		c.q.Push(sendFailedQueue, buf)
	}
}
