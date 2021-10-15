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


package devicemapper

import (
	"fmt"
	"github.com/pkg/errors"
	"polardb-sms/pkg/agent/utils"
	smslog "polardb-sms/pkg/log"
	"strings"
	"time"
)

const (
	FileNotFound = "No such file or directory"
)

type scsi struct {
}

/*
  scsi device operate 函数簇
  ## scanFcDevice() error
  ## scanIscsiDevice() error
  ## rescanScsiDevice(volumeName string, options SshOptions) error
*/
func (s *scsi) ScanDeviceFcHost() error {
	fcScanCmd := fmt.Sprintf("for x in `ls /sys/class/fc_host`; do " +
		"echo \"- - -\" > /sys/class/scsi_host/$x/scan; done")
	out, stderr, err := utils.Exec(fcScanCmd, 20 * time.Second)
	if err != nil {
		smslog.Errorf("exec command %s stderr %s err %s", fcScanCmd, stderr, err.Error())
		return fmt.Errorf("failed scan fc host: %v", err)
	}

	if strings.Contains(out, FileNotFound) {
		smslog.Warn("Not found fc host path, skip scan...")
		return nil
	}

	smslog.Info("successfully scan fc host")
	return nil
}

func (s *scsi) ScanDeviceIscsiHost() error {
	iScsiScanCmd := fmt.Sprintf("for x in `ls /sys/class/iscsi_host`; do " +
		"echo \"- - -\" > /sys/class/scsi_host/$x/scan; done")

	out, stderr, err := utils.Exec(iScsiScanCmd, 20 * time.Second)
	if err != nil {
		smslog.Errorf("exec command %s stderr %s err %s", iScsiScanCmd, stderr, err.Error())
		return fmt.Errorf("failed scan iscsi host: %v", err)
	}

	if strings.Contains(out, FileNotFound) {
		smslog.Warn("Not found iscsi host path, skip scan...")
		return nil
	}

	smslog.Info("successfully scan iscsi host")
	return nil
}

func (s *scsi) ScsiDeviceRescan() error {
	/*
		# 重新扫描特定的 SCSI ClusterDevice, echo 1 > /sys/block/$DEVICE/device/rescan 用sda, sdb, sdc等替换$DEVICE
	*/
	rescanCmd := fmt.Sprintf("for x in $(multipath -l|grep ' active'|awk '{print $(NF-4)}'); do " +
		"echo 1 > /sys/block/$x/device/rescan; done")
	_, stderr, err := utils.Exec(rescanCmd, 60 * time.Second)
	if err != nil {
		smslog.Errorf("exec command %s stderr %s err %s", rescanCmd, stderr, err.Error())
		return errors.Wrap(err, stderr)
	}
	smslog.Info("successfully scsi device rescan")

	return nil
}

func newScsi() *scsi {
	return &scsi{}
}
