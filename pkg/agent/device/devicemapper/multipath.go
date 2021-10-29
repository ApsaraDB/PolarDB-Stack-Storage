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
	"polardb-sms/pkg/device"
	smslog "polardb-sms/pkg/log"
	"strings"
)

type multiPath struct {
	WWID   string `json:"wwid"`
	Alias  string `json:"alias,omitempty"`
	Device string `json:"device,omitempty"`
	Vendor string `json:"vendor,omitempty"`
	Size   string `json:"size"`
}

/*
   multipath operate 函数簇
   ## MultiPathCreate(deviceName, wwid string) error
   ## MultiPathResize(deviceName string) error
   ## MultiPathReload(deviceName string) error
   ## MultiPathRemove(deviceName string) error
*/
func (m *multiPath) MultiPathCreate(deviceName, wwid string) error {
	mpCreateCmd := fmt.Sprintf(`cat > /etc/multipath/conf.d/%s.conf << EOF
multipaths {
  multipath {
    wwid     %s
    alias    %s
  }
}
EOF`, deviceName, wwid, deviceName)
	_, stderr, err := utils.ExecCommand(mpCreateCmd, utils.CmdDefaultTimeout)
	if err != nil {
		smslog.Errorf("exec command %s stderr %s err %s", mpCreateCmd, stderr, err.Error())
		return errors.Wrap(err, stderr)
	}
	smslog.Infof("successfully create multipath config %s", deviceName)

	return nil
}

func (m *multiPath) MultiPathSearch(deviceName string) bool {
	mpSearchCmd := fmt.Sprintf("ls /dev/mapper/%s", deviceName)
	_, stderr, err := utils.ExecCommand(mpSearchCmd, utils.CmdDefaultTimeout)
	if err != nil {
		smslog.Errorf("exec command %s stderr %s err %s", mpSearchCmd, stderr, err.Error())
		return false
	}
	smslog.Infof("successfully search multipath config %s", deviceName)

	return true
}

func (m *multiPath) MultiPathDetail() ([]multiPath, error) {
	/*
		## 36e00084100ee7ec97e9547a30000227f dm-56 HUAWEI  ,XSG1            ,size=1.0T features='0' hwhandler='0' wp=rw
		## 360050767088080a26800000000006811 dm-28 ALIBABA ,MCS             ,size=210G features='1 queue_if_no_path' hwhandler='0' wp=rw
		## pv-36e00084100ee7ec97b8f295400001fb0 (36e00084100ee7ec97b8f295400001fb0) dm-61 HUAWEI  ,XSG1            ,size=220G features='0' hwhandler='0' wp=rw
		## pv-360050767088080a2680000000000678a (360050767088080a2680000000000678a) dm-25 ALIBABA ,MCS             ,size=510G features='1 queue_if_no_path' hwhandler='0' wp=rw
	*/
	mpDetailCmd := `multipath -ll|grep size= -B 1|grep -v "\-\-"|awk '{if(NR%2==0){printf $0 "\n"}else{printf "%s,",$0}}'|grep features`
	out, stderr, err := utils.ExecCommand(mpDetailCmd, utils.CmdDefaultTimeout)
	if err != nil {
		smslog.Errorf("exec command %s stderr %s err %s", mpDetailCmd, stderr, err.Error())
		return nil, fmt.Errorf("failed exec multipath detail: %v", err)
	}
	smslog.Infof("successfully exec multipath detail")

	var multipaths []multiPath
	if len(out) != 0 {
		for _, line := range strings.Split(out, device.NewLineSign) {
			var multipath multiPath
			fields := strings.Fields(line)
			if len(fields) == 0 {
				break
			}
			if strings.Contains(fields[1], "(") && strings.Contains(fields[1], ")") {
				multipath.Alias = fields[0]
				multipath.WWID = fields[1][1 : len(fields[1])-1]
				multipath.Device = fields[2]
				multipath.Vendor = fields[3]
				multipath.Size = strings.Split(fields[5], ",size=")[1]
			} else {
				multipath.Alias = fields[0]
				multipath.WWID = fields[0]
				multipath.Device = fields[1]
				multipath.Vendor = fields[2]
				multipath.Size = strings.Split(fields[4], ",size=")[1]
			}
			multipaths = append(multipaths, multipath)
		}
	}
	return multipaths, nil
}

func (m *multiPath) MultiPathResize(deviceName string) error {
	/*
		# Resize your multipath device by running the multipathd resize command
		multipathd resize map ${vdisk_name}
	*/
	mpResizeCmd := fmt.Sprintf("multipathd resize map %s", deviceName)
	_, stderr, err := utils.ExecCommand(mpResizeCmd, utils.CmdDefaultTimeout)
	if err != nil {
		smslog.Errorf("exec command %s stderr %s err %s", mpResizeCmd, stderr, err.Error())
	}
	smslog.Infof("successfully exec multipathd resize map %s", deviceName)

	return nil
}

func (m *multiPath) MultiPathReload() error {
	/*
		# reload multipathd
		systemctl reload multipathd
	*/
	mpReloadCmd := fmt.Sprintf("systemctl reload multipathd")
	_, stderr, err := utils.ExecCommand(mpReloadCmd, utils.CmdDefaultTimeout)
	if err != nil {
		smslog.Errorf("exec command %s stderr %s err %s", mpReloadCmd, stderr, err.Error())
	}
	smslog.Info("successfully systemctl reload multipathd")

	return nil
}

func (m *multiPath) MultiPathRemove(deviceName string) error {
	mpRemoveCmd := fmt.Sprintf("rm -rf /etc/multipath/conf.d/%s.conf", deviceName)
	_, stderr, err := utils.ExecCommand(mpRemoveCmd, utils.CmdDefaultTimeout)
	if err != nil {
		smslog.Errorf("exec command %s stderr %s err %s", mpRemoveCmd, stderr, err.Error())
	}
	smslog.Infof("successfully remove multipath config %s", deviceName)

	return nil
}

func newMultiPath() *multiPath {
	return &multiPath{}
}
