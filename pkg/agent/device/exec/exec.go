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


/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package exec

import (
	"fmt"
	"net"
	"os/exec"
	smslog "polardb-sms/pkg/log"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const (
	GiB = 1024 * 1024 * 1024
	MiB = 1024 * 1024
)

func EXEC(cmd string, tags ...string) (string, error) {
	cnt := getTracingID()
	tag := fmt.Sprintf("%d|%s", cnt, strings.Join(tags, "|"))

	start := time.Now()
	smslog.Debugf("EXEC[%s]Command-stdin[%s]", tag, cmd)
	outInfo, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		smslog.Debugf("EXEC[%s]Command-stdin[%s] exec err %s， outInfo %s", tag, cmd, err.Error(), outInfo)
		return "", fmt.Errorf("EXEC[%s]Command-stderr[%v]", tag, err)
	}
	smslog.Debugf("EXEC[%s]Command-stdout[%s], cost time: %v(s)", tag, outInfo, time.Since(start).Seconds())

	return string(outInfo), nil
}

// cmdRunnerCounter CmdRunnerCounter
var cmdRunnerCounter uint64 = 0

func getTracingID() uint64 {
	return atomic.AddUint64(&cmdRunnerCounter, 1)
}

func GetDevicePath(volumeId string) string {
	return fmt.Sprintf("/dev/mapper/%s", volumeId)
}

func BytesToGiB(volumeSizeBytes int64) int64 {
	return volumeSizeBytes / GiB
}

func GiBToBytes(volumeSizeGiB int64) int64 {
	return volumeSizeGiB * GiB
}

func NodeIPToScsiPrKey(node string) string {
	ip := net.ParseIP(node)
	if len(ip) == 0 {
		return ""
	}
	i := int(ip[12]) * 16777216
	i += int(ip[13]) * 65536
	i += int(ip[14]) * 256
	i += int(ip[15])
	return fmt.Sprintf("%#x", i)
}

func ScsiPrKeyToNodeIP(key string) string {
	key = strings.Replace(key, "0x", "", 1)
	i, err := strconv.ParseInt(key, 16, 64)
	if err != nil {
		panic(err)
	}
	d := byte(i % 256)
	i = i / 256
	c := byte(i % 256)
	i = i / 256
	b := byte(i % 256)
	i = i / 256
	a := byte(i)
	return net.IPv4(a, b, c, d).To4().String()
}
