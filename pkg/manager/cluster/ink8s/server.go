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


package ink8s

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/cluster/arp"
	"sync"
	"time"
)

const StorageNamespace string = "kube-system"

type InK8sServerConfig struct {
	Vip             string
	DistributedLock string
	ClientSet       kubernetes.Interface
	Ip              string
	Id              string
	StopCh          <-chan struct{}
}

type InK8sServer struct {
	InK8sServerConfig
	Lock         sync.RWMutex
	innerRunners map[string]RunnerWithLeader
}

type RunnerWithLeader interface {
	Run()
	Stop()
	Identify() string
}

func NewServer(c *InK8sServerConfig) *InK8sServer {
	return &InK8sServer{
		InK8sServerConfig: *c,
		innerRunners:      make(map[string]RunnerWithLeader),
	}
}

func (s *InK8sServer) AddRunner(runner RunnerWithLeader) {
	s.innerRunners[runner.Identify()] = runner
}

func (s *InK8sServer) Run() {
	smslog.Debugf("k8s lease nodeId is %s", s.Id)
	defer smslog.LogPanic()
	localId := s.Id
	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      s.DistributedLock,
			Namespace: StorageNamespace,
		},
		Client: s.ClientSet.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: localId,
		},
	}

	leaderelection.RunOrDie(context.Background(), leaderelection.LeaderElectionConfig{
		Name:            s.DistributedLock,
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   60 * time.Second, //租约时间
		RenewDeadline:   15 * time.Second, //更新租约的
		RetryPeriod:     5 * time.Second,  //非leader节点重试时间
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				smslog.Debug("OnStartedLeading")
				s.runForLeader(ctx)
			},
			OnStoppedLeading: func() {
				smslog.Debug("OnStoppedLeading")
				s.runStopLeader()
			},
			OnNewLeader: func(identity string) {
				smslog.Debugf("OnNewLeader %s localId %s", identity, localId)
				if identity != localId {
					s.runStopLeader()
				}
			},
		},
	})
	<-s.StopCh
}

func (s *InK8sServer) runForLeader(ctx context.Context) {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	smslog.Infof("runForLeader: current response id %s ip %s become leader", s.Id, s.Ip)
	err := arp.ARPSendGratuitousByIp(s.Vip, s.Ip)
	if err != nil {
		smslog.Infof("error update arp for vip %s, ip %s err %s", s.Vip, s.Ip, err.Error())
	}
	for _, runner := range s.innerRunners {
		go runner.Run()
	}
}

func (s *InK8sServer) runStopLeader() {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	for _, runner := range s.innerRunners {
		runner.Stop()
	}
}
