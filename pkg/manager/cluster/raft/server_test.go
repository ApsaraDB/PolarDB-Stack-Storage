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
	"github.com/stretchr/testify/assert"
	"testing"
)

var server *RaftServer

func TestConnectTos(t *testing.T) {
	go setup()
	testCase := assert.New(t)
	client, err := ConnectTo("127.0.0.1", "9009")
	if err != nil {
		println("Error xxxxxxxxxxxxx: %s", err.Error())
		testCase.Error(err)
		return
	}
	messageResp, err := client.OnHeartbeat(context.Background(), &HeartbeatRequest{
		From: 0,
		To:   1,
		Term: 0,
	})
	if err != nil {
		println("yyyyyyy %s", err.Error())
		return
	}
	println("xxxxxxx" + messageResp.String())

	voteResp, err := client.OnRequestVote(context.Background(), &VoteRequest{
		From:        0,
		To:          1,
		Term:        0,
		CandidateId: 0,
	})
	if err != nil {
		println("yyyyyyy %s", err.Error())
		return
	}
	println("xxxxxxx" + voteResp.String())

}

func setup() {
	server = NewServer(nil)
	server.Run()
}
