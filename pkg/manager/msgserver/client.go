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

package msgserver

import (
	"bufio"
	"fmt"
	"io"
	"net"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/network"
	"time"

	"polardb-sms/pkg/network/message"
)

const (
	SleepBase   = 1
	SleepFactor = 2
	SleepMax    = 60
)

//manager侧为client
type SmsClient struct {
	network.SmsServerConfig
	conn   *network.SmsConnection
	RecvCh chan *message.SmsMessage
	SendCh chan *message.SmsMessage
	ready  bool
}

func NewClient(conf *network.SmsServerConfig, ch chan *message.SmsMessage) *SmsClient {
	return &SmsClient{
		SmsServerConfig: *conf,
		RecvCh:          ch,
		SendCh:          make(chan *message.SmsMessage),
	}
}

func (c *SmsClient) Run() {
	//connect to response
	defer smslog.LogPanic()
	c.connect()
	go c.receive()
	go c.send()
	smslog.Infof("request started %s:%s", c.Ip, c.Port)
	c.ready = true
}

func (c *SmsClient) String() string {
	return fmt.Sprintf("[%s:%s]", c.Ip, c.Port)
}

func (c *SmsClient) send() {
	defer smslog.LogPanic()
	for {
		msg := <-c.SendCh
		err := c.conn.Send(msg)
		if err != nil {
			smslog.WithContext(msg.Head.TraceContext).Error(err.Error())
		} else {
			smslog.WithContext(msg.Head.TraceContext).Infof("client: %s already sent message: %s", c.String(), msg.Head.MsgId)
		}
	}
}

func (c *SmsClient) receive() {
	defer smslog.LogPanic()
	for {
		msg, err := c.conn.Receive()
		if err == nil {
			//smslog.WithContext(msg.Head.TraceContext).Debugf("received message %v", msg)
			c.RecvCh <- msg
			continue
		}
		smslog.Errorf("received message err %s", err.Error())
		if err == io.EOF {
			c.reconnect()
		}
	}
}

func (c *SmsClient) reconnect() {
	smslog.Infof("start to reconnect to %s", c.Ip)
	c.connect()
}

func (c *SmsClient) connect() {
	connFun := func() error {
		tcpAddr, err := net.ResolveTCPAddr("tcp4", c.Ip+":"+c.Port)
		if err != nil {
			smslog.Fatalf("Invalid tcp address for response : %s:%s, err %v", c.Ip, c.Port, err)
			return err
		}
		conn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			smslog.Errorf("could not dial TCP %s for response: %v", tcpAddr, err)
			return err
		}
		smslog.Infof("connecting dial TCP %s for response", tcpAddr)

		_ = conn.SetKeepAlive(true)
		if c.conn != nil {
			c.conn.Close()
		}
		c.conn = &network.SmsConnection{
			Conn: conn,
			Receiver: &network.SmsMessageReceiver{
				Reader: bufio.NewReader(conn),
			},
			Sender: &network.SmsMessageSender{
				Writer: bufio.NewWriter(conn),
			},
		}
		return nil
	}
	sleepTime := SleepBase
	for {
		err := connFun()
		if err == nil {
			break
		} else {
			time.Sleep(time.Duration(sleepTime) * time.Second)
		}
		sleepTime = sleepTime * SleepFactor
		if sleepTime > SleepMax {
			sleepTime = SleepMax
		}
	}
}
