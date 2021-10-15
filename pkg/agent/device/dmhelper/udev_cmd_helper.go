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
	"polardb-sms/pkg/agent/utils"
	"polardb-sms/pkg/device"
	smslog "polardb-sms/pkg/log"
	"strings"
)

/**
  Example:

  #udevadm info /dev/dm-29
  P: /devices/virtual/block/dm-29
  N: dm-29
  L: 10
  S: disk/by-id/dm-name-3600507670880813290000000000003f4
  S: disk/by-id/dm-uuid-mpath-3600507670880813290000000000003f4
  S: mapper/3600507670880813290000000000003f4
  E: DEVLINKS=/dev/disk/by-id/dm-name-3600507670880813290000000000003f4 /dev/disk/by-id/dm-uuid-mpath-3600507670880813290000000000003f4 /dev/mapper/3600507670880813290000000000003f4
  E: DEVNAME=/dev/dm-29
  E: DEVPATH=/devices/virtual/block/dm-29
  E: DEVTYPE=disk
  E: DM_ACTIVATION=0
  E: DM_MULTIPATH_TIMESTAMP=1606680175
  E: DM_NAME=3600507670880813290000000000003f4
  E: DM_SUBSYSTEM_UDEV_FLAG0=1
  E: DM_SUSPENDED=0
  E: DM_UDEV_DISABLE_LIBRARY_FALLBACK_FLAG=1
  E: DM_UDEV_PRIMARY_SOURCE_FLAG=1
  E: DM_UDEV_RULES_VSN=2
  E: DM_UUID=mpath-3600507670880813290000000000003f4
  E: MAJOR=252
  E: MINOR=29
  E: MPATH_SBIN_PATH=/sbin
  E: SUBSYSTEM=block
  E: TAGS=:systemd:
  E: USEC_INITIALIZED=810430

  解释：
  P: Paths
  N: DmName
  L: Priority
  E: Property
*/
type UDevParam struct {
	Path     string
	Name     string
	Symlink  []string
	Priority string
	Property map[string]string
}

func getUDevParam(name string) (*UDevParam, error) {
	var cmd string
	if strings.Index(name, "dm-") == 0 {
		cmd = fmt.Sprintf("udevadm info /dev/%s", name)
	} else if strings.Index(name, "dm-") == 5 || strings.Contains(name, "/dev/") {
		cmd = fmt.Sprintf("udevadm info %s", name)
	} else {
		cmd = fmt.Sprintf("udevadm info /dev/mapper/%s", name)
	}
	stdout, stderr, err := utils.ExecCommand(cmd, utils.CmdDefaultTimeout)
	if err != nil || stderr != "" {
		return nil, fmt.Errorf("exec '%s' failed, stderr: %v, err: %v", cmd, stderr, err)
	}
	smslog.Infof("exec: %s, stdout: %s", cmd, stdout)

	var uDevAdm UDevParam
	for _, row := range strings.Split(stdout, device.NewLineSign) {
		fields := strings.Fields(row)
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "P:":
			uDevAdm.Path = fields[1]
		case "N:":
			uDevAdm.Name = fields[1]
		case "S:":
			uDevAdm.Symlink = append(uDevAdm.Symlink, fields[1])
		case "E:":
			cells := strings.Split(fields[1], device.SemicolonSign)
			if uDevAdm.Property == nil {
				uDevAdm.Property = make(map[string]string)
			}
			uDevAdm.Property[cells[0]] = cells[1]
		case "L:":
			uDevAdm.Priority = fields[1]
		}
	}
	return &uDevAdm, nil
}
