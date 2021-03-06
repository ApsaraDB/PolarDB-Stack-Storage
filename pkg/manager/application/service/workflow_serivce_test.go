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

package service

import (
	"fmt"
	"polardb-sms/pkg/common"
	"polardb-sms/pkg/manager/application/view"
	"testing"
	"time"
)

func TestWflEngine_Submit(t *testing.T) {

	v := &view.ClusterLvFormatRequest{
		VolumeId: "21",
	}
	v.VolumeName = "21"
	fmt.Print(v.VolumeName)
	fmt.Print(v.FsType != common.NoFs)
	fmt.Print(v.FsType == common.NoFs)
}

func TestCloseChan(t *testing.T) {
	stopCh := make(chan struct{})
	testFun := func() {
		for {
			select {
			case <-stopCh:
				println("stoped")
				return
			default:
				println("hello")
				time.Sleep(time.Second)
			}
		}
	}
	go testFun()
	go testFun()
	time.Sleep(5 * time.Second)
	close(stopCh)

	time.Sleep(2 * time.Second)
	go testFun()
	time.Sleep(5 * time.Second)
}
