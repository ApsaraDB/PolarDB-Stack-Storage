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
	"math/rand"
	"sync"
	"time"
)

type roleType string

const (
	Leader    roleType = "leader"
	Follower  roleType = "follower"
	Candidate roleType = "candidate"
)

type raftState struct {
	term   int64
	vote   int64
	commit int64
}

const (
	OneTick          int = 20 //20ms
	ElectionTimeout  int = 30 //30 * OneTick
	HeartbeatTimeout int = 5  //5 * OneTick
)

const None uint64 = 0

type raftNode struct {
	mu       sync.Mutex
	id       uint64
	term     uint64
	voteFor  uint64
	voteCnt  int
	role     roleType
	leaderId uint64

	heartbeatElapsed int
	heartbeatTimeout int
	heartbeatFlag    bool

	electionElapsed           int
	electionTimeout           int
	randomizedElectionTimeout int
	//heartbeat chan
	hbCh chan *Message
	//just for voteFor message chan
	recvCh chan *Message
	sendCh chan *Message
	peers  map[uint64]bool
}

func NewRaftNode(c *raftConfig, rc chan *Message, sc chan *Message) *raftNode {
	r := &raftNode{
		id:               c.id,
		role:             "",
		leaderId:         None,
		heartbeatTimeout: c.heartbeatTick,
		electionTimeout:  c.electionTick,
		peers:            make(map[uint64]bool),
		recvCh:           rc,
		sendCh:           sc,
		hbCh:             make(chan *Message),
	}

	//TODO set handler loop
	return r
}

/**
节点刚启动，进入follower状态，同时创建一个超时时间在150-300毫秒之间的选举超时定时器。
*/
func (node *raftNode) Run() {
	node.role = Follower
	node.term = 0
	node.voteFor = 1
	ticker := time.NewTicker(5 * time.Microsecond)
	defer ticker.Stop()
	smslog.Infof("RaftNode %v start to run", node.id)
	for {
		<-ticker.C
		switch node.role {
		case Follower:
			node.runAsFollower()
		case Candidate:
			node.runAsCandidate()
		case Leader:
			node.runAsLeader()
		default:
			smslog.Fatalf("Wrong node role: %v", node.role)
			return
		}
	}
}

/**
follower状态节点主循环：
  如果收到leader节点心跳：
    心跳标志位置1
  如果选举超时到期：
    没有收到leader节点心跳：
      任期号term+1，换到candidate状态。
    如果收到leader节点心跳：
      心跳标志位置空
  如果收到选举消息：
 	如果最近一段时间还有收到来自leader节点的心跳消息：
    	拒绝该请求，返回
    如果当前没有给任何节点投票过 或者 消息的任期号大于当前任期号：
      投票给该节点
    否则：
      拒绝投票给该节点
*/
func (node *raftNode) runAsFollower() {
	smslog.Infof("%d Start to run as follower", node.id)
	for {
		select {
		case <-node.hbCh: //received heartbeat
			node.heartbeatFlag = true
		case msg := <-node.recvCh: //received voteFor message??
			if node.heartbeatFlag == true && msg.Term <= node.term {
				//拒绝请求
				return
			}
			if node.voteFor == 0 || msg.Term > node.term {
				smslog.Infof("follower %d  vote yes to %d for message %v", node.id, msg.From, msg)
				node.voteTo(msg.From, msg.Term, true)
			} else {
				smslog.Infof("follower %d  vote no to %d for message %v", node.id, msg.From, msg)
				node.voteTo(msg.From, msg.Term, false)
			}
		case <-time.After(time.Duration(500+rand.Intn(300)) * time.Millisecond):
			smslog.Infof("follower %d election time out", node.id)
			if node.heartbeatFlag == false {
				node.role = Candidate
				return
			} else {
				node.heartbeatFlag = false
			}
		}
	}
}

/**
candidate状态节点主循环：
  向集群中其他节点发送RequestVote请求，请求中带上当前任期号term
  收到AppendEntries消息：
    如果该消息的任期号 >= 本节点任期号term：
      说明已经有leader，切换到follower状态
    否则：
      拒绝该消息
  收到其他节点应答RequestVote消息：
    如果数量超过集群半数以上，切换到leader状态

  如果选举超时到期：
    term+1，进行下一次的选举
*/
func (node *raftNode) runAsCandidate() {
	smslog.Infof("%d Start to run as candidate", node.id)
	node.term++
	node.voteFor = node.id
	node.voteCnt = 1
	go node.broadcastVoteRequest()
	for {
		select {
		case <-time.After(time.Duration(rand.Intn(200)+500) * time.Millisecond):
			smslog.Infof("candidate %v time out", node.id)
			return
		case msg := <-node.hbCh: //valid message and work as follower
			smslog.Infof("candidate %d received hb %v", node.id, msg)
			if msg.Term >= node.term {
				node.role = Follower
				return
			}
		case msg := <-node.recvCh: //received vote response
			smslog.Infof("candidate %d received msg %v", node.id, msg)
			if msg.Type == TypeVoteResp && msg.Grant == true {
				node.voteCnt++
				if node.voteCnt > (len(node.peers)+1+1)/2 {
					node.role = Leader
					return
				}
			}
		}
	}
}

func (node *raftNode) runAsLeader() {
	smslog.Infof("%d Start to run as leader", node.id)
	//TODO add status check and change
	for {
		node.broadcastHeartbeat()
		time.Sleep(time.Duration(node.heartbeatTimeout*OneTick) * time.Microsecond)
	}
}

func (node *raftNode) sendHeartbeat(to uint64) {
	node.sendCh <- &Message{
		Type: TypeHeartbeatReq,
		From: node.id,
		To:   to,
		Term: node.term,
	}
}

func (node *raftNode) broadcastHeartbeat() {
	for peerId, exit := range node.peers {
		if exit {
			node.sendHeartbeat(peerId)
		}
	}
}

func (node *raftNode) sendVoteRequest(to uint64) {
	node.sendCh <- &Message{
		Type: TypeVoteReq,
		From: node.id,
		To:   to,
		Term: node.term,
	}
}
func (node *raftNode) broadcastVoteRequest() {
	for peerId, exist := range node.peers {
		if exist {
			node.sendVoteRequest(peerId)
		}
	}
}

func (node *raftNode) onHeartbeat(req *Message) {
	node.hbCh <- req
}

func (node *raftNode) onVote(req *Message) {
	node.recvCh <- req
}

func (node *raftNode) connectToPeer(pId uint64) {
	node.peers[pId] = true
}

func (node *raftNode) voteTo(peer uint64, term uint64, granted bool) {
	if node.term < term {
		node.term = term
	}
	node.sendCh <- &Message{
		Type:  TypeVoteResp,
		From:  node.id,
		To:    peer,
		Term:  node.term,
		Grant: granted,
	}
}