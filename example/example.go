package main

import (
	"time"

	"github.com/eastany/apnsd/models/client"
)

type Client struct {
	ntc *client.NotifiClient
	fdc *client.FeedbackClient
}

func do() {
	clt := &Client{
		ntc: client.NewNotifiClient("eastany.com:9001"),
		fdc: client.NewFeedbackClient("eastany.com:9002"),
	}
	if clt.ntc == nil {
		return
	}
	i := 0
	for i < 100000 {
		clt.ntc.Send(nil)
		i++
	}
}

func main() {
	go do()
	go do()
	go do()
	time.Sleep(time.Second * 100)
}
