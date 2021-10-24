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
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/config"
	"polardb-sms/pkg/manager/msgserver/handler"
	"polardb-sms/pkg/network"
	"polardb-sms/pkg/network/message"
	"sync"
)

type MessageServerConfig struct {
	serverConfs map[string]network.SmsServerConfig
}

var MsgServer *MessageServer

type MessageServer struct {
	MessageServerConfig
	agentMap   map[string]*SmsClient
	msgService *handler.RespMsgHandleService
	recvCh     chan *message.SmsMessage
	waitingMsg map[string]chan interface{}
	sync.Mutex
}

func NewMessageServer(agentMap map[string]config.Node) *MessageServer {
	MsgServer = &MessageServer{
		msgService: handler.NewRespMsgHandleService(),
		agentMap:   make(map[string]*SmsClient),
		waitingMsg: make(map[string]chan interface{}),
		recvCh:     make(chan *message.SmsMessage),
	}
	for key, val := range agentMap {
		MsgServer.agentMap[key] = NewClient(&network.SmsServerConfig{
			Ip:   val.Ip,
			Port: val.Port,
		}, MsgServer.recvCh)
	}
	return MsgServer
}

func (s *MessageServer) Run() {
	for name, c := range s.agentMap {
		go c.Run()
		smslog.Infof("Connect to agent %s started", name)
	}
	go s.receive()
}

func (s *MessageServer) SendMessageTo(name string, msg *message.SmsMessage, ch chan interface{}) {
	c, ok := s.agentMap[name]
	if !ok {
		//TODO add new agent
		smslog.WithContext(msg.Head.TraceContext).Errorf("can not find the agent client for %s, agentMap %v", name, s.agentMap)
		return
	}
	s.Lock()
	defer s.Unlock()
	s.waitingMsg[msg.Head.MsgId] = ch
	c.SendCh <- msg
}

func (s *MessageServer) receive() {
	defer smslog.LogPanic()
	var sg sync.WaitGroup
	for {
		msg := <-s.recvCh
		sg.Add(1)
		go func() {
			defer smslog.LogPanic()
			defer sg.Done()
			defer func() {
				s.Lock()
				delete(s.waitingMsg, msg.Head.AckMsgId)
				s.Unlock()
			}()

			ch, ok := s.waitingMsg[msg.Head.AckMsgId]
			if !ok {
				smslog.WithContext(msg.Head.TraceContext).Errorf("received msg %v, but cannot find the callback", msg)
				return
			}
			result, err := s.msgService.Handle(msg)
			if err != nil {
				smslog.WithContext(msg.Head.TraceContext).Errorf("handle msg error: %v", err)
				return
			}
			ch <- result
			smslog.Debugf("finished process msg %s", msg)
		}()
		sg.Wait()
	}
}
