package notification

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/rand"
	"strconv"
	"time"
)

const pushCommandValue = 2

const MaxPayloadSizeBytes = 2048

const IdentifierUbound = 9999

const (
	deviceTokenItemid            = 1
	payloadItemid                = 2
	notificationIdentifierItemid = 3
	expirationDateItemid         = 4
	priorityItemid               = 5
	deviceTokenLength            = 32
	notificationIdentifierLength = 4
	expirationDateLength         = 4
	priorityLength               = 1
)

type Payload struct {
	Alert            interface{} `json:"alert,omitempty"`
	Badge            int         `json:"badge,omitempty"`
	Sound            string      `json:"sound,omitempty"`
	ContentAvailable int         `json:"content-available,omitempty"`
	Category         string      `json:"category,omitempty"`
}

func NewPayload() *Payload {
	return new(Payload)
}

type AlertDictionary struct {
	Title        string   `json:"title,omitempty"`
	Body         string   `json:"body,omitempty"`
	TitleLocKey  string   `json:"title-loc-key,omitempty"`
	TitleLocArgs []string `json:"title-loc-args,omitempty"`
	ActionLocKey string   `json:"action-loc-key,omitempty"`
	LocKey       string   `json:"loc-key,omitempty"`
	LocArgs      []string `json:"loc-args,omitempty"`
	LaunchImage  string   `json:"launch-image,omitempty"`
}

func NewAlertDictionary() *AlertDictionary {
	return new(AlertDictionary)
}

type Notification struct {
	Identifier  int32
	Expiry      uint32
	DeviceToken string
	payload     map[string]interface{}
	Priority    uint8
}

func NewNotification() (nf *Notification) {
	nf = new(Notification)
	nf.payload = make(map[string]interface{})
	nf.Identifier = rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(IdentifierUbound)
	nf.Priority = 10
	return
}

func (nf *Notification) AddPayload(p *Payload) {
	if p.Badge == 0 {
		p.Badge = -1
	}
	nf.Set("aps", p)
}

func (nf *Notification) Get(key string) interface{} {
	return nf.payload[key]
}

func (nf *Notification) Set(key string, value interface{}) {
	nf.payload[key] = value
}

func (nf *Notification) PayloadJSON() ([]byte, error) {
	return json.Marshal(nf.payload)
}

func (nf *Notification) PayloadString() (string, error) {
	j, err := nf.PayloadJSON()
	return string(j), err
}

func (nf *Notification) ToBytes() ([]byte, error) {
	token, err := hex.DecodeString(nf.DeviceToken)
	if err != nil {
		return nil, err
	}
	if len(token) != deviceTokenLength {
		return nil, errors.New("device token has incorrect length")
	}
	payload, err := nf.PayloadJSON()
	if err != nil {
		return nil, err
	}
	if len(payload) > MaxPayloadSizeBytes {
		return nil, errors.New("payload is larger than the " + strconv.Itoa(MaxPayloadSizeBytes) + " byte limit")
	}
	frameBuffer := new(bytes.Buffer)
	binary.Write(frameBuffer, binary.BigEndian, uint8(deviceTokenItemid))
	binary.Write(frameBuffer, binary.BigEndian, uint16(deviceTokenLength))
	binary.Write(frameBuffer, binary.BigEndian, token)
	binary.Write(frameBuffer, binary.BigEndian, uint8(payloadItemid))
	binary.Write(frameBuffer, binary.BigEndian, uint16(len(payload)))
	binary.Write(frameBuffer, binary.BigEndian, payload)
	binary.Write(frameBuffer, binary.BigEndian, uint8(notificationIdentifierItemid))
	binary.Write(frameBuffer, binary.BigEndian, uint16(notificationIdentifierLength))
	binary.Write(frameBuffer, binary.BigEndian, nf.Identifier)
	binary.Write(frameBuffer, binary.BigEndian, uint8(expirationDateItemid))
	binary.Write(frameBuffer, binary.BigEndian, uint16(expirationDateLength))
	binary.Write(frameBuffer, binary.BigEndian, nf.Expiry)
	binary.Write(frameBuffer, binary.BigEndian, uint8(priorityItemid))
	binary.Write(frameBuffer, binary.BigEndian, uint16(priorityLength))
	binary.Write(frameBuffer, binary.BigEndian, nf.Priority)
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, uint8(pushCommandValue))
	binary.Write(buffer, binary.BigEndian, uint32(frameBuffer.Len()))
	binary.Write(buffer, binary.BigEndian, frameBuffer.Bytes())
	return buffer.Bytes(), nil
}

func ParseNotificationToken(data []byte) string {
	return hex.EncodeToString(data[7:11])
}
