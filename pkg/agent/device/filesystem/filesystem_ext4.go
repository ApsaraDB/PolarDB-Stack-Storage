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


package filesystem

import (
	"fmt"
	"polardb-sms/pkg/agent/device/dmhelper"
	"polardb-sms/pkg/agent/device/exec"
	"polardb-sms/pkg/agent/utils"
	"polardb-sms/pkg/common"
	"strconv"
	"strings"
	smslog "polardb-sms/pkg/log"
	log "polardb-sms/pkg/log"
	"time"
)

const (
	// FSTypeExt2 represents the ext2 filesystem type
	FSTypeExt2 = "ext2"
	// FSTypeExt3 represents the ext3 filesystem type
	FSTypeExt3 = "ext3"
	// FSTypeExt4 represents the ext4 filesystem type
	FSTypeExt4 = "ext4"
	// FSTypeXfs represents te xfs filesystem type
	FSTypeXfs = "xfs"

	// default file system type to be used when it is not provided
	DefaultFsType = FSTypeExt4
)

var _ Filesystem = &Ext4{}

type Ext4 struct {
}

//TODO
func (e *Ext4) BrowseFilesystem(deviceName string) (int64, error) {
	return 0, nil
}

func (e *Ext4) FormatFilesystem(deviceName string) error {
	/*
		[root@r03.dbm-02 ~]$mkfs.ext4 -F -m0 /dev/mapper/pv-36e00084100ee7ec96ad2f05d00000cb2
		mke2fs 1.42.9 (28-Dec-2013)
		Discarding device blocks: done
		Filesystem label=
		OS type: Linux
		Block size=4096 (log=2)
		Fragment size=4096 (log=2)
		Stride=0 blocks, Stripe width=0 blocks
		13762560 inodes, 55050240 blocks
		0 blocks (0.00%) reserved for the super user
		First data block=0
		Maximum filesystem blocks=2204106752
		1680 block groups
		32768 blocks per group, 32768 fragments per group
		8192 inodes per group
		Superblock backups stored on blocks:
		        32768, 98304, 163840, 229376, 294912, 819200, 884736, 1605632, 2654208,
		        4096000, 7962624, 11239424, 20480000, 23887872

		Allocating group tables: done
		Writing inode tables: done
		Creating journal (32768 blocks): done
		Writing superblocks and filesystem accounting information: done
	*/

	log.Infof("VolumeInfo %s appears to be unformatted, attempting to format as type: %q", deviceName, DefaultFsType)
	devicePath, err := common.GetDevicePath(deviceName)
	if err != nil {
		return err
	}
	ext4MkfsCmd := fmt.Sprintf("mkfs.%s -F -D -m0 %s", DefaultFsType, devicePath)
	blockDevBytes, err := dmhelper.GetBlockDevSize(deviceName)
	if err != nil {
		return err
	}
	reqSizeIn100GiB := blockDevBytes / (100 * 1024 * 1024 * 1024)
	stdout, stderr, err := utils.ExecCommand(ext4MkfsCmd, time.Duration(20 + 10 * reqSizeIn100GiB) * time.Second)
	if err != nil {
		smslog.Debugf("mkfs %s failed, stdout: %s, stderr: %s, err: %s", deviceName, stdout, stderr, err)
		return fmt.Errorf("failed exec command %s err %s", ext4MkfsCmd, err)
	}
	log.Infof("VolumeInfo successfully formatted (mkfs): %s - %s", DefaultFsType, deviceName)

	return nil
}

func (e *Ext4) ExpandFilesystem(deviceName string, expandCapacity int64, originCapacity int64) error {
	/*
	   resize2fs ${device_path}
	   # first resize2fs output:
	   ## resize2fs 1.42.9 (28-Dec-2013)
	   ## Filesystem at /dev/mapper/pvc-be649fe0-25f4-11ea-8abb-50af732f4b8f is mounted on /root/tmp; on-line resizing required
	   ## old_desc_blocks = 1, new_desc_blocks = 1
	   ## The filesystem on /dev/mapper/pvc-be649fe0-25f4-11ea-8abb-50af732f4b8f is now 524288 blocks long.

	   # next resize2fs output:
	   ## resize2fs 1.42.9 (28-Dec-2013)
	   ## The filesystem is already 524288 blocks long.  Nothing to do!
	*/

	log.Infof("VolumeInfo %s appears to be expanded", deviceName)
	devicePath, err := common.GetDevicePath(deviceName)
	if err != nil {
		return err
	}
	resize2fsCmd := fmt.Sprintf("resize2fs %s", devicePath)
	outInfo, stderr, err := utils.ExecCommand(resize2fsCmd, utils.CmdDefaultTimeout)
	if err != nil {
		smslog.Debugf("resize2fs %s failed, stdout: %s, stderr: %s, err: %s", devicePath, outInfo, stderr, err)
		if strings.Contains(err.Error(), "Nothing to do") {
			//The filesystem is already 208404480 blocks long.  Nothing to do!
			latestSize := calcSize(err.Error(), "already", "block")
			if latestSize == expandCapacity {
				return nil
			}
		}
		return fmt.Errorf("failed exec command %s err %s", resize2fsCmd, err)
	}
	// The filesystem on /dev/mapper/polar-00074xb679652s-1190 is now 268697600 blocks long.
	actualGiB := calcSize(outInfo, "now", "blocks")
	if actualGiB != exec.BytesToGiB(expandCapacity) {
		err = fmt.Errorf("could not resize2fs %s, result: [%s], actual: [%dGiB]", devicePath, outInfo, actualGiB)
	}
	log.Infof("successfully exec resize2fs %s", deviceName)

	return nil
}

func calcSize(out, starting, ending string) int64 {
	s := strings.Index(out, starting)
	if s < 0 {
		return -1
	}
	s += len(starting)

	e := strings.Index(out[s:], ending)
	if e < 0 {
		return -1
	}

	blocks := out[s : s+e]
	var value, err = strconv.Atoi(strings.TrimSpace(blocks))
	if err != nil {
		return -1
	}

	return int64(value * 4 / 1024 / 1024)
}

func getExt4Cmd(fsType, source string) string {
	return fmt.Sprintf("mkfs.%s -F -D -m0 %s", fsType, source)
}

func getResize2fsCmd(devicePath string) string {
	return fmt.Sprintf("resize2fs %s", devicePath)
}

func getE2fsckCmd(devicePath string) string {
	return fmt.Sprintf("e2fsck -fy %s", devicePath)
}

func NewExt4() Filesystem {
	return &Ext4{}
}
