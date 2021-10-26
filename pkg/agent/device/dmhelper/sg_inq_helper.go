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

package dmhelper

import (
	"fmt"
	"github.com/pkg/errors"
	"polardb-sms/pkg/agent/utils"
	smslog "polardb-sms/pkg/log"
	"strings"
)

func getSqInqParams(name string) (string, error) {
	cmd := fmt.Sprintf("sg_inq /dev/mapper/%s | grep \"serial number\"", name)
	stdout, stderr, err := utils.ExecCommand(cmd, utils.CmdDefaultTimeout)
	if err != nil {
		smslog.Errorf("sg_inq %s err %s stderr %s", name, err.Error(), stderr)
		return "", errors.Wrap(err, stderr)
	}
	serialNumber := strings.Trim(stdout, "\n")
	if serialNumber != "" {
		return serialNumber, nil
	}
	return "", fmt.Errorf("can not find serial number")
}
