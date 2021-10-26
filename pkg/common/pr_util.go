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
	"net"
	"strings"
)

const hexDigit = "0123456789abcdef"

func hexString(b []byte) string {
	s := make([]byte, len(b)*2)
	for i, tn := range b {
		s[i*2], s[i*2+1] = hexDigit[tn>>4], hexDigit[tn&0xf]
	}
	return "0x" + string(s)
}

func IpV4ToPrKey(ipStr string) string {
	ip := net.ParseIP(ipStr)
	return hexString(ip.To4())
}

func PrKeyToIpV4(hexStr string) string {
	if strings.HasPrefix(hexStr, "0x") {
		hexStr = strings.TrimPrefix(hexStr, "0x")
	}
	if hexStr == "" || len(hexStr) < 7 {
		return ""
	}
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}
	ipBytes := []byte(hexStr)
	s := make([]byte, 4)
	for i, _ := range s {
		s[i] = byte(strings.Index(hexDigit, string(ipBytes[i*2]))<<4 + strings.Index(hexDigit, string(ipBytes[i*2+1]))&0xf)
	}
	return net.IPv4(s[0], s[1], s[2], s[3]).String()
}
