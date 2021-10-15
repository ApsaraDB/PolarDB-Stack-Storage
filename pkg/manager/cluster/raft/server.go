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


package raft

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"math/rand"
	"net"
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/cluster/arp"
	"strconv"
	"sync"
	"time"
)

const (
	DefaultProtocol = "tcp"
	DefaultPort     = ":9009"
)

type MessageType int

const (
	TypeVoteReq MessageType = iota
	TypeVoteResp
	TypeHeartbeatReq
	TypeHeartbeatResp
	TypeBecomeLeader
)

type NodeConfig struct {
	Id   string
	Ip   string
	Port string
}

type RaftServerConfig struct {
	NodeConfig
	Vip   string
	Peers []*NodeConfig
}

type RaftServer struct {
	RaftServerConfig
	rpcServer   *grpc.Server
	node        *raftNode
	peersMap    map[uint64]string
	PeerClients map[string]RaftServiceClient
	listener    net.Listener
	wg          sync.WaitGroup
	quit        chan interface{}
	ready       <-chan interface{}
	sendCh      chan *Message
	recvCh      chan *Message
	vRespCh     chan *VoteResponse
}

type Message struct {
	Type  MessageType
	From  uint64
	To    uint64
	Term  uint64
	Grant bool
}

func ip2id(ip string, port string) uint64 {
	b := net.ParseIP(ip).To4()
	if b == nil {
		rand.Seed(time.Now().UnixNano())
		return uint64(rand.Int63n(10000))
	}
	ipInt := uint64(b[3]) | uint64(b[2])<<8 | uint64(b[1])<<16 | uint64(b[0])<<24
	p, err := strconv.Atoi(port)
	if err != nil {
		return ipInt
	}
	return ipInt<<8 | uint64(p)
}

func NewServer(c *RaftServerConfig) *RaftServer {
	s := RaftServer{
		RaftServerConfig: *c,
	}
	s.peersMap = make(map[uint64]string)
	s.PeerClients = make(map[string]RaftServiceClient)
	s.wg = sync.WaitGroup{}
	s.quit = make(chan interface{})
	s.sendCh = make(chan *Message)
	s.recvCh = make(chan *Message)
	s.vRespCh = make(chan *VoteResponse)
	s.node = NewRaftNode(&raftConfig{
		id: ip2id(s.Ip, s.Port),
	}, s.sendCh, s.recvCh)
	return &s
}

func (s *RaftServer) Run() {
	var err error
	s.listener, err = net.Listen(DefaultProtocol, fmt.Sprintf("%s:%s", s.Ip, s.Port))
	if err != nil {
		panic("Fatal error when listen: " + err.Error())
	}
	s.rpcServer = grpc.NewServer()
	RegisterRaftServiceServer(s.rpcServer, s)
	reflection.Register(s.rpcServer)
	smslog.Infof("start to run response: %v", s.RaftServerConfig)
	go s.rpcServer.Serve(s.listener)
	go s.node.Run()
	go func() {
		for {
			select {
			case msg := <-s.recvCh:
				s.process(msg)
			}
		}
	}()
	for _, peer := range s.Peers {
		go common.RunWithRetry(3, 10*time.Millisecond, func(retryTimes int) error {
			return s.ConnectToPeer(peer.Id, peer.Ip, peer.Port)
		})
	}
}

func (s *RaftServer) process(msg *Message) {
	switch msg.Type {
	case TypeHeartbeatReq:
		_, _ = s.Heartbeat(&HeartbeatRequest{
			From: msg.From,
			To:   msg.To,
			Term: msg.Term,
		})
	case TypeVoteReq:
		resp, err := s.RequestVote(&VoteRequest{
			From:        msg.From,
			To:          msg.To,
			Term:        msg.Term,
			CandidateId: msg.From,
		})
		if err != nil {
			s.sendCh <- &Message{
				Type:  TypeVoteResp,
				From:  msg.To,
				To:    msg.From,
				Term:  msg.Term,
				Grant: false,
			}
		} else {
			s.sendCh <- &Message{
				Type:  TypeVoteResp,
				From:  resp.From,
				To:    resp.To,
				Term:  resp.Term,
				Grant: resp.Granted,
			}
		}

	case TypeVoteResp:
		s.vRespCh <- &VoteResponse{
			From:    msg.From,
			To:      msg.To,
			Term:    msg.Term,
			Granted: msg.Grant,
		}
	case TypeHeartbeatResp:
	//do nothing future maybe
	case TypeBecomeLeader:
		//TODO arp broadcast
		smslog.Infof("current response id %s ip %s become leader", s.Id, s.Ip)
		_ = arp.ARPSendGratuitousByIp(s.Vip, s.Ip)
	}
}

//TODO fix this
func (s *RaftServer) ShutDown() {
}

func (s *RaftServer) ConnectToPeer(peerId string, ip string, port string) error {
	client, err := ConnectTo(ip, port)
	if err != nil {
		smslog.Errorf("node: %s connect to peer %s ip %s port %s error: %s", s.Id, peerId, ip, port, err.Error())
		return err
	}
	s.PeerClients[peerId] = client
	pId := ip2id(ip, port)
	s.peersMap[pId] = peerId
	s.node.connectToPeer(pId)
	smslog.Infof("node %v successfully connect to peer %s ip %s  port %s", s.Id, peerId, ip, port)
	return nil
}

//TODO fix this
func (s *RaftServer) DisconnectPeer(peerId string) error {
	return nil
}

func (s *RaftServer) OnHeartbeat(ctx context.Context, req *HeartbeatRequest) (*HeartbeatResponse, error) {
	s.node.onHeartbeat(&Message{
		Type: TypeHeartbeatReq,
		From: req.From,
		To:   req.To,
		Term: req.Term,
	})
	return &HeartbeatResponse{
		From:   s.node.id,
		To:     req.From,
		Term:   s.node.term,
		Reject: false,
	}, nil
}

func (s *RaftServer) OnRequestVote(ctx context.Context, req *VoteRequest) (*VoteResponse, error) {
	s.node.onVote(&Message{
		Type: TypeVoteReq,
		From: req.From,
		To:   req.To,
		Term: req.Term,
	})
	select {
	case resp := <-s.vRespCh:
		smslog.Infof("raftServer %s received vote response %v", s.Id, resp)
		return resp, nil
	case <-time.After(200 * time.Millisecond):
		smslog.Infof("response %s OnRequestVote timeout", s.Id)
		return &VoteResponse{
			From:    s.node.id,
			To:      req.From,
			Term:    s.node.term,
			Granted: false,
		}, nil
	}
}

func (s *RaftServer) getClient(id uint64) (RaftServiceClient, error) {
	idStr, ok := s.peersMap[id]
	if ok {
		c, ok := s.PeerClients[idStr]
		if ok && c != nil {
			return c, nil
		}
	}
	return nil, fmt.Errorf("can not find peer request for id %d", id)
}

func (s *RaftServer) Heartbeat(req *HeartbeatRequest) (*HeartbeatResponse, error) {
	client, err := s.getClient(req.To)
	if err != nil {
		smslog.Infof("Can not find the request for hb request: %v, error: %s", req, err.Error())
		return nil, err
	}
	return client.OnHeartbeat(context.Background(), req)
}
func (s *RaftServer) RequestVote(req *VoteRequest) (*VoteResponse, error) {
	client, err := s.getClient(req.To)
	if err != nil {
		smslog.Infof("Can not find the request for hb request: %v, error: %s", req, err.Error())
		return nil, err
	}
	return client.OnRequestVote(context.Background(), req)
}



