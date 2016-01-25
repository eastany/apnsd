package feedback

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"
)

const (
	DEFAULT_TOKEN_LEN = 32
)

type Feedback struct {
	Ts          uint32
	TokenLen    uint16
	DeviceToken string
}

func NewFeedback(token string) *Feedback {
	return &Feedback{
		Ts:          uint32(time.Now().Unix()),
		TokenLen:    DEFAULT_TOKEN_LEN,
		DeviceToken: token,
	}
}

func (f *Feedback) ToBytes() ([]byte, error) {
	token, err := hex.DecodeString(f.DeviceToken)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, f.Ts)
	binary.Write(buffer, binary.BigEndian, f.TokenLen)
	binary.Write(buffer, binary.BigEndian, token)
	return buffer.Bytes(), nil
}
