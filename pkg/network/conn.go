/*
*Copyright (c) 2019-2021, Alibaba Group Holding Limited;
*Licensed under the Apache License, Version 2.0 (the "License");
*you may not use this file except in compliance with the License.
*You may obtain a copy of the License at

*   http://www.apache.org/licenses/LICENSE-2.0

*Unless required by applicable law or agreed to in writing, software
*distributed under the License is distributed on an "AS IS" BASIS,
*WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*See the License for the specific language governing permissions and
*limitations under the License.
 */

package network

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"google.golang.org/protobuf/proto"
	"io"
	"net"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/network/message"
	"sync"
	"time"
)

type SmsServerConfig struct {
	Ip   string
	Port string
}

const (
	SmsMessageEnd byte = 0x1d
)

type SmsMessageReceiver struct {
	Reader *bufio.Reader
}

func (r *SmsMessageReceiver) receive() (*message.SmsMessage, error) {
	msgLine, err := r.Reader.ReadString(SmsMessageEnd)
	if err != nil {
		smslog.Errorf("received message err %s", err.Error())
		if err == io.EOF {
			//TODO Handle conn is closed
			return nil, io.EOF
		}
		return nil, err
	}
	msg, err := Parse(msgLine)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

type SmsMessageSender struct {
	Writer *bufio.Writer
}

func (s *SmsMessageSender) Send(msg *message.SmsMessage) {
	output, err := proto.Marshal(msg)
	if err != nil {
		//TODO Handle
		smslog.WithContext(msg.Head.TraceContext).Errorf("could not send message: %v", err)
		return
	}
	sEnc := base64.StdEncoding.EncodeToString(output)
	output = append([]byte(sEnc), SmsMessageEnd)
	length, err := s.Writer.Write(output)
	if err = s.Writer.Flush(); err != nil {
		smslog.WithContext(msg.Head.TraceContext).Errorf("could not write tcp: %v", err)
		return
	}
	smslog.Infof("successfully send message: [%v], length: %d", msg, length)
}

func Parse(line string) (*message.SmsMessage, error) {
	msg := &message.SmsMessage{}
	contents := []byte(line)
	contents = contents[:len(contents)-1]
	sDec, err := base64.StdEncoding.DecodeString(string(contents))
	if err != nil {
		smslog.Errorf("failed decode string %s to message", string(sDec))
		return nil, err
	}
	err = proto.Unmarshal(sDec, msg)
	if err != nil {
		smslog.Errorf("failed parse string %s to message", line)
		return nil, err
	}
	return msg, nil
}

type SmsConnection struct {
	sync.Mutex
	Conn     net.Conn
	Receiver *SmsMessageReceiver
	Sender   *SmsMessageSender
}

func (c *SmsConnection) Receive() (*message.SmsMessage, error) {
	return c.Receiver.receive()
}

func (c *SmsConnection) Send(msg *message.SmsMessage) error {
	c.Lock()
	defer c.Unlock()
	sendFinishCh := make(chan bool)
	go func() {
		c.Sender.Send(msg)
		sendFinishCh <- true
	}()
	select {
	case <-sendFinishCh:
	case <-time.After(1 * time.Second):
		return fmt.Errorf("send msg %s timeout", msg.Head.MsgId)
	}
	return nil
}

func (c *SmsConnection) Close() {
	if c.Conn != nil {
		_ = c.Conn.Close()
	}
}
