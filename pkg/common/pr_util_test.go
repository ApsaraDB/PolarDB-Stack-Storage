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


package common

import (
	"testing"
	"time"
)

func TestIpToPrKey(t *testing.T) {
	//testCase := assert.New(t)
	ipStr := "0.2.1.0"
	key := IpV4ToPrKey(ipStr)
	println(key)

	ipStr2 := PrKeyToIpV4(key)
	println(ipStr2)
}

func TestPrKeyToIpV4(t *testing.T) {
	ipStr := "10.238.11.193"
	key := IpV4ToPrKey(ipStr)
	println(key)
	ipStr1 := PrKeyToIpV4("0xac20e19f")
	println(ipStr1)

	ipStr2 := PrKeyToIpV4(key)
	println(ipStr2)

	ipStr3 := PrKeyToIpV4("0xc6134203")
	println(ipStr3)
}

func TestIpV4ToPrKey(t *testing.T) {
	stopCh := make(chan struct{})
	for i := 0; i < 5; i++ {
		go func(stopCh chan struct{}, idx int) {
			for {
				select {
				case <-stopCh:
					println("stop %d", idx)
					return
				case <-time.After(1 * time.Second):
					println("1 sec")
				}
			}
		}(stopCh, i)
	}

	time.Sleep(20 * time.Second)
	close(stopCh)
	time.Sleep(10 * time.Second)
}
