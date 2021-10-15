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
	"io"
	"math/rand"
	"net"
	"polardb-sms/pkg/agent/msgserver/handler"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/network"
	"strings"
	"time"
)

var server *MessageServer

type MessageServer struct {
	nodeId     string
	nodeIp     string
	port       string
	clients    map[string]*network.SmsConnection
	msgService handler.ReqMsgHandlerService
	listener   net.Listener
}

func NewMessageServer(nodeId, nodeIp, port string) *MessageServer {
	server = &MessageServer{
		nodeId:     nodeId,
		nodeIp:     nodeIp,
		port:       port,
		clients:    make(map[string]*network.SmsConnection, 0),
		msgService: handler.NewReqMsgHandlerService(nodeIp),
	}
	return server
}

func PickOneClientIp() string {
	if server == nil {
		return ""
	}
	if len(server.clients) == 0 {
		return ""
	}
	randIdx := rand.Intn(len(server.clients))
	var add = 0
	for addr, _ := range server.clients {
		if add == randIdx {
			hostPortArray := strings.Split(addr, ":")
			return hostPortArray[0]
		}
		add++
	}
	return ""
}

func (s *MessageServer) Run() {
	defer smslog.LogPanic()
	for {
		l, err := net.Listen("tcp", s.nodeIp+":"+s.port)
		if err != nil {
			smslog.Error("error when listen:  " + err.Error())
			time.Sleep(2 * time.Second)
			continue
		}
		s.listener = l
		break
	}
	smslog.Infof("Agent server %s:%s starting...", s.nodeIp, s.port)
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			smslog.Errorf("fatal error when listen %s", err.Error())
			continue
		}
		clientAddr := conn.RemoteAddr().String()
		existConn, ok := s.clients[clientAddr]
		if ok && existConn != nil {
			smslog.Debugf("Close exist connection %v", existConn)
			existConn.Close()
		}
		s.clients[clientAddr] = &network.SmsConnection{
			Conn: conn,
			Receiver: &network.SmsMessageReceiver{
				Reader: bufio.NewReader(conn),
			},
			Sender: &network.SmsMessageSender{
				Writer: bufio.NewWriter(conn),
			},
		}
		go s.receiveAndHandleMessage(s.clients[clientAddr])
		smslog.Infof("accept connection from request %s", clientAddr)
	}
}

func (s *MessageServer) Shutdown() {
	for _, conn := range s.clients {
		_ = conn.Conn.Close()
	}
	err := s.listener.Close()
	if err != nil {
		smslog.Fatal("Server forced to shutdown:", err)
		return
	}
	smslog.Infof("Server exiting")
}

//单线程模式
func (s *MessageServer) receiveAndHandleMessage(conn *network.SmsConnection) {
	defer smslog.LogPanic()
	for {
		msg, err := conn.Receive()
		if err != nil {
			if err == io.EOF {
				smslog.Errorf("conn %v is closed, received message err: %v", conn.Conn.RemoteAddr().String(), err)
				delete(s.clients, conn.Conn.RemoteAddr().String())
				break
			}
			smslog.Errorf("received message err: %v", err)
			continue
		}
		smslog.Infof("received message: [%v]", msg)
		go func() {
			defer smslog.LogPanic()
			result, err := s.msgService.Process(msg)
			if err == nil {
				err = conn.Send(result)
				if err != nil {
					smslog.WithContext(result.Head.TraceContext).Errorf("send msg err %s", err.Error())
				}
			} else {
				smslog.Errorf("Handle message error: %s", err.Error())
			}
		}()
	}
}
