package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"sync/atomic"
	"time"

	tlsconn "github.com/eastany/apnsd/common/connection"
	nt "github.com/eastany/apnsd/common/notification"
	"github.com/eastany/apnsd/common/queue"
)

const (
	_sendFailedQueue = "notification_send_failed_q"
)

type NotifiClient struct {
	Conn *tlsconn.TlsConnection
	q    queue.Queue

	sended     uint64
	sendFailed uint64
	resended   uint64
}

func NewNotifiClient(host string) (nc *NotifiClient) {
	root, err := ioutil.ReadFile("ca.pem")
	if err != nil {
		fmt.Println(err)
		return
	}
	cert2_b, err := ioutil.ReadFile("cert2.pem")
	if err != nil {
		fmt.Println(err)
		return
	}
	priv2_b, err := ioutil.ReadFile("cert2.key")
	if err != nil {
		fmt.Println(err)
		return
	}
	priv2, err := x509.ParsePKCS1PrivateKey(priv2_b)
	if err != nil {
		fmt.Println(err)
		return
	}
	rootCA, err := x509.ParseCertificate(root)
	if err != nil {
		fmt.Println(err)
		return
	}
	cert := tls.Certificate{
		Certificate: [][]byte{cert2_b},
		PrivateKey:  priv2,
	}

	conn := tlsconn.NewTlsConnWithCert(rootCA, cert, host)
	nc = &NotifiClient{Conn: conn}
	nc.Conn.Connect()
	go nc.redial()
	return
}

func (c *NotifiClient) Send(n *nt.Notification) {
	data := nt.MockNotification()
	_, err := c.send(data)
	if err != nil {
		fmt.Println(err)
		c.q.Push(_sendFailedQueue, data)
		atomic.AddUint64(&c.sendFailed, 1)
		return
	}
	atomic.AddUint64(&c.sended, 1)
}

func (c *NotifiClient) reSendLoop() {
	for {
		data := c.q.Pop(_sendFailedQueue)
		if data != nil && len(data) > 1 {
			atomic.AddUint64(&c.resended, 1)
			c.send(data)
		} else {
			time.Sleep(time.Second * 1)
		}
	}
}

func (c *NotifiClient) send(data []byte) (int, error) {
	return c.Conn.WriteAll(data)
}

func (c *NotifiClient) redial() {
	for {
		c.Conn.Connect()
		time.Sleep(time.Second * 1)
	}
}
