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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePrRegister(t *testing.T) {
	keyInfo := `  
  PR generation=0x85, 	15 registered reservation keys follow:
    0x1
    0x1
    0x1
    0x1
    0x1
    0x1
    0x1
    0x2
    0x2
    0x2
    0x2
    0x2
    0x2
    0x2
    0x2
`
	reserveInfo := `
  PR generation=0x85, Reservation follows:
   Key = 0x1
  scope = LU_SCOPE, type = Write Exclusive
`

	pr, err := ParsePrRegister(keyInfo, reserveInfo)
	assert.NoError(t, err)

	assert.Equal(t, "0x1", pr.ReservationKey)
	assert.Equal(t, 7, pr.Keys["0x1"])
	assert.Equal(t, 8, pr.Keys["0x2"])
	assert.Equal(t, "Write Exclusive", pr.ReservationType)

	parts := strings.SplitN("a-b-c", "-", 2)
	fmt.Println(parts[1])
}

func TestParseCapabilities(t *testing.T) {
	info := `
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
`

	capabilities := ParseCapabilities(info)

	assert.True(t, capabilities.Support(PRC_WR_EX))
	assert.False(t, capabilities.Support(PRC_PTPL_A))
}
