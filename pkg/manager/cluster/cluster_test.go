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


package cluster

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/cluster/ink8s"
	"polardb-sms/pkg/manager/cluster/raft"
	"testing"
	"time"
)

func bootstrapByK8s(t *Topology, lockName string, clientSet kubernetes.Interface) {
	for nodeName := range t.Topology {
		n := t.Topology[nodeName]
		n.s = ink8s.NewServer(&ink8s.InK8sServerConfig{
			Vip:             t.Vip,
			DistributedLock: lockName,
			ClientSet:       clientSet,
			Ip:              n.Ip,
			Id:              n.Id,
		})
		go n.s.Run()
	}
}

func bootstrapByRaw(t *Topology) {
	for nodeName := range t.Topology {
		n := t.Topology[nodeName]
		n.s = raft.NewServer(&raft.RaftServerConfig{
			NodeConfig: raft.NodeConfig{
				Id:   n.Id,
				Ip:   n.Ip,
				Port: n.Port,
			},
			Vip: t.Vip,
		})
		go n.s.Run()
	}
	for nodeName := range t.Topology {
		n := t.Topology[nodeName]
		for _, peer := range n.peers {
			_ = n.s.(*raft.RaftServer).ConnectToPeer(peer.Id, peer.Ip, peer.Port)
		}
	}
}

func bootstrap() *Topology {
	c := &Config{
		NodeCnt:  3,
		NodesMap: make(map[string]NodeConfig),
	}
	for i := 0; i < c.NodeCnt; i++ {
		n := NodeConfig{
			Id:   fmt.Sprint(i),
			Ip:   "127.0.0.1",
			Name: "response-" + fmt.Sprint(i),
			Port: fmt.Sprint(9000 + i),
		}
		c.NodesMap[n.Name] = n
	}
	topo := NewCluster(c)
	bootstrapByRaw(topo)
	return topo
}

func TestServer_OnHeartbeat(t *testing.T) {
	testCase := assert.New(t)
	topo := bootstrap()
	idx := 2
	randServer := topo.Topology["response-"+fmt.Sprint(idx)].s
	result, err := randServer.(*raft.RaftServer).PeerClients[fmt.Sprint(3-idx)].OnHeartbeat(context.Background(), &raft.HeartbeatRequest{
		From: 1,
		To:   1,
		Term: 1,
	})
	smslog.Infof("result is : %v", result)
	testCase.NoError(err)
	testCase.Equal(result.Reject, false)
	time.Sleep(time.Second * 1)
}
