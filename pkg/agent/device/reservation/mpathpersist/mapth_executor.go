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


package mpathpersist

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/ngaut/log"
	"polardb-sms/pkg/agent/utils"
)

const (
	MPLogLevel     = 2
	DefaultTimeout = 5 * time.Second
)

type PersistentReserve struct {
	Keys            map[string]int
	Generation      string
	ReservationKey  string
	ReservationType string
}

/**
[root@r03.dbm-01 ~]$sg_persist -c --in /dev/sder
  ALIBABA   MCS               0000
  Peripheral device type: disk
Report capabilities response:
  Compatible Reservation Handling(CRH): 1
  Specify Initiator Ports Capable(SIP_C): 0
  All Target Ports Capable(ATP_C): 0
  Persist Through Power Loss Capable(PTPL_C): 1
  Type Mask Valid(TMV): 1
  Allow Commands: 0
  Persist Through Power Loss Active(PTPL_A): 0
    Support indicated in Type mask:
      Write Exclusive, all registrants: 1
      Exclusive Access, registrants only: 1
      Write Exclusive, registrants only: 1
      Exclusive Access: 1
      Write Exclusive: 1
      Exclusive Access, all registrants: 1

RLR_C (Replace Lost Reservation Capable) bit
1 Indicates that the device server supports the REPLACE LOST RESERVATION service action in the PERSISTENT RESERVE OUT command.
0 Indicates that the device server does not support the REPLACE LOST RESERVATION service action in the PERSISTENT RESERVE OUT command. If set to 0 then the device server shall not terminate any commands with CHECK CONDITION status with the sense key set to DATA PROTECT and the additional sense code set to PERSISTENT RESERVATION INFORMATION LOST as described in SPC-5.

CRH (Compatible Reservation Handling) field
1 A Compatible Reservation Handling (CRH) bit set to one indicates that the device server supports the exceptions to the RESERVE and
RELEASE commands described in SPC-5.
0 A CRH bit set to zero indicates that RESERVE(6) command, RESERVE(10) command, RELEASE(6) command, and RELEASE(10) command are processed as defined in SPC-5.

SIP_C (Specify Initiator Ports Capable) bit
1 A Specify Initiator Ports Capable (SIP_C) bit set to one indicates that the device server supports the SPEC_I_PT bit in the PERSISTENT
RESERVE OUT command parameter data (see 3.14.3).
0 An SIP_C bit set to zero indicates that the device server does not support the SPEC_I_PT bit in the PERSISTENT RESERVE OUT command parameter data.

ATP_C (Target Ports Capable) bit
1 An All Target Ports Capable (ATP_C) bit set to one indicates that the device server supports the ALL_TG_PT bit in the PERSISTENT
RESERVE OUT command parameter data.
0 An ATP_C bit set to zero indicates that the device server does not support the ALL_TG_PT bit in the PERSISTENT RESERVE OUT command parameter data.
 SCSI Commands Reference Manual, Rev. J 120
www.seagate.com Direct Access Block commands (SPC-5 and SBC-4)

PTPL_C (Persist Through Power Loss Capable) bit
1 A Persist Through Power Loss Capable (PTPL_C) bit set to one indicates that the device server supports the persist through power loss
capability see SPC-5 for persistent reservations and the APTPL bit in the PERSISTENT RESERVE OUT command parameter data.
0 An PTPL_C bit set to zero indicates that the device server does not support the persist through power loss capability.

TMV (Type Mask Valid) bit
1 A Type Mask Valid (TMV) bit set to one indicates that the PERSISTENT RESERVATION TYPE MASK field contains a bit map indicating which
persistent reservation types are supported by the device server.
0 A TMV bit set to zero indicates that the PERSISTENT RESERVATION TYPE MASK field shall be ignored.

PTPL_A (Persist Through Power Loss Activated) bit
1 A Persist Through Power Loss Activated (PTPL_A) bit set to one indicates that the persist through power loss capability is activated (see
SPC-5).
0 A PTPL_A bit set to zero indicates that the persist through power loss capability is not activated.

*/
type PRCapability string

const (
	PRC_CRH      PRCapability = "Compatible Reservation Handling(CRH)"
	PRC_SIP_C                 = "Specify Initiator Ports Capable(SIP_C)"
	PRC_ATP_C                 = "All Target Ports Capable(ATP_C)"
	PRC_PTPL_C                = "Persist Through Power Loss Capable(PTPL_C)"
	PRC_PTPL_A                = "Persist Through Power Loss Active(PTPL_A)"
	PRC_EX_AC_AR              = "Exclusive Access, all registrants"
	PRC_WR_EX                 = "Write Exclusive, all registrants"
)

type PRCapabilities map[PRCapability]interface{}

func (p PRCapabilities) String() string {
	var ret string
	for k, v := range p {
		ret = ret + fmt.Sprintf("%s:%s;", k, v)
	}
	return ret
}

func (p PRCapabilities) Support(capability PRCapability) bool {
	if v, ok := p[capability]; ok {
		return v == "1"
	}
	return false
}

func NewPersistentReserve() *PersistentReserve {
	return &PersistentReserve{
		Keys: make(map[string]int),
	}
}

// 查询能力需要使用物理设备
func ReportCapabilities(device string) (*PRCapabilities, error) {
	cmd := fmt.Sprintf("sg_persist -c --in %s", device)

	outInfo, errInfo, err := utils.ExecCommand(cmd, DefaultTimeout)
	if err != nil || errInfo != "" {
		return nil, fmt.Errorf("exec cmd: %s err stdout: %s, stderr: %s, err: %s", cmd, outInfo, errInfo, err)
	}

	return ParseCapabilities(outInfo), nil
}

func IsDeviceExist(device string) (bool, error) {
	cmd := fmt.Sprintf("ls %s", device)

	outInfo, errInfo, err := utils.ExecCommand(cmd, DefaultTimeout)
	if err != nil || errInfo != "" {
		if strings.Contains(errInfo, "No such file or directory") {
			log.Infof("device %s not exist ", device)
			return false, nil
		}
		return false, fmt.Errorf("stdout: %s, stderr: %s, err: %s", outInfo, errInfo, err)
	}
	return true, nil
}

func getPrInfo(device string) (*PersistentReserve, error) {
	var (
		cmd string
	)
	cmd = fmt.Sprintf("mpathpersist -v %d --in -k %s", MPLogLevel, device)
	keyOut, keyErr, err := utils.ExecCommand(cmd, DefaultTimeout)
	if err != nil || keyErr != "" {
		return nil, fmt.Errorf("failed to query pr key, stdout: %s, stderr: %s, err: %s", keyOut, keyErr, err)
	}

	cmd = fmt.Sprintf("mpathpersist -v %d --in -r %s", MPLogLevel, device)
	reserveOut, reserveErr, err := utils.ExecCommand(cmd, DefaultTimeout)
	if err != nil || reserveErr != "" {
		return nil, fmt.Errorf("failed to query pr reservation, stdout: %s, stderr: %s, err: %s", reserveOut, reserveErr, err)
	}

	return ParsePrRegister(keyOut, reserveOut)
}

func GetPrInfo(device string) (*PersistentReserve, error) {
	pr, err := getPrInfo(device)
	if err == ErrGenerationChanged {
		time.Sleep(1 * time.Second)
		return getPrInfo(device)
	}
	return pr, err
}

var ErrGenerationChanged = fmt.Errorf("key and reservation generation not equal, may be has already changed, please retry")

/**
[root@r010a0001.cloud.a0.amtest8 /root]
#mpathpersist --in -k /dev/mapper/polar-000vx9h0w684dc-3355
  PR generation=0x85, 	7 registered reservation keys follow:
    0x1
    0x1
    0x1
    0x1
    0x1
    0x1
    0x1

[root@r010a0001.cloud.a0.amtest8 /root]
#mpathpersist --in -r /dev/mapper/polar-000vx9h0w684dc-3355
  PR generation=0x85, Reservation follows:
   Key = 0x1
  scope = LU_SCOPE, type = Write Exclusive

*/
func ParsePrRegister(keyInfo, reserveInfo string) (*PersistentReserve, error) {
	var (
		pr                = NewPersistentReserve()
		keyGeneration     string
		reserveGeneration string
	)
	for _, line := range strings.Split(keyInfo, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "PR generation") {
			parts := strings.SplitN(line, ",", 2)
			keyGeneration = strings.TrimSpace(strings.Split(parts[0], "=")[1])
		} else if strings.HasPrefix(line, "0x") {
			count, ok := pr.Keys[line]
			if !ok {
				count = 0
			}
			pr.Keys[line] = count + 1
		}
	}

	for _, line := range strings.Split(reserveInfo, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "PR generation") {
			parts := strings.SplitN(line, ",", 2)
			reserveGeneration = strings.TrimSpace(strings.Split(parts[0], "=")[1])
		} else if strings.HasPrefix(line, "Key") {
			parts := strings.SplitN(line, "=", 2)
			pr.ReservationKey = strings.TrimSpace(parts[1])
		} else if strings.HasPrefix(line, "scope") {
			parts := strings.SplitN(line, ",", 2)
			pr.ReservationType = strings.TrimSpace(strings.Split(parts[1], "=")[1])
		}
	}

	if keyGeneration != reserveGeneration {
		return nil, ErrGenerationChanged
	} else {
		pr.Generation = keyGeneration
	}

	return pr, nil
}

func ParseCapabilities(info string) *PRCapabilities {
	capabilities := PRCapabilities{}
	for _, line := range strings.Split(info, "\n") {
		line = strings.TrimSpace(line)
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}

		k := strings.TrimSpace(line[:idx])
		v := strings.TrimSpace(line[idx+1:])

		if len(k) == 0 || len(v) == 0 {
			continue
		}

		capabilities[PRCapability(k)] = v
	}
	return &capabilities
}

func NodeToScsiPrKey(node string) string {
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

func ScsiPrKeyToNode(key string) string {
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
