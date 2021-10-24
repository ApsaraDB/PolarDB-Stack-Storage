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
	"github.com/pkg/errors"
	"polardb-sms/pkg/agent/device/dmhelper"
	"polardb-sms/pkg/agent/device/exec"
	"polardb-sms/pkg/agent/utils"
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
	"strconv"
	"strings"
	"time"
)

// constants of pfs relevant information
const (
	// pfs场景， 不能创建小于 10G 的盘
	PfsMinGiB = 10
	// pfs mkfs 磁盘大小的万分之一，超过 1T 的创建请求自动拆分，以 1T 为步长来扩容
	PfsMaxGiB = 1 * 1024 * 1024 * 1024 * 1024
	// 只能用 10 的倍数, 比如 10G 扩容到 20G, 然后 -o 1 -n 2
	PfsChunkNum = 10
	// pfs nall num
	PfsNallNum = 4
	// pfs 格式化成功输出
	PfsGrowFs = "pfs growfs succeeds!"
)

// constants of ssh stdout filter out separator
const (
	DefaultRowSplitStr      = "|"
	DefaultRowColonStr      = ":"
	DefaultLineSplitStr     = "\n"
	DefaultRowSpaceSplitStr = " "
)

type PfsOptions struct {
	command     string
	oldChunkNum int
	newChunkNum int
}

var _ Filesystem = &Pfs{}

type Pfs struct {
}

func (p *Pfs) BrowseFilesystem(deviceName string) (int64, error) {
	/*
		#打印PBD mapper_xhuqvqm-polar-0007z2bs1p490h-rwo-5225文件系统的元数据信息
		## e.g., output
		## >pfs -C disk info mapper_xhuqvqm-polar-0007z2bs1p490h-rwo-5225
		## Blktag Info:
		## (0)allocnode: id 0, shift 0, nchild=3, nall 7680, nfree 5283, next 0
		## Direntry Info:
		## (0)allocnode: id 0, shift 0, nchild=3, nall 6144, nfree 4030, next 0
		## Inode Info:
		## (0)allocnode: id 0, shift 0, nchild=3, nall 6144, nfree 4030, next 0

		## Blktag Info nchild是chunk个数, 总容量 = 7680 * 4M(30720MB), 空闲 block 容量 = 5283 * 4MB(21132MB)
		## nall: 元数据总个数
		## nfree: 元数据空闲个数
	*/
	pbdName, err := common.GetPBDName(deviceName)
	if err != nil {
		return 0, errors.Wrap(err, fmt.Sprintf("failed get DevicePath by deviceName %s", deviceName))
	}

	pfsInfoCmd := getPfsCmd(pbdName, PfsOptions{command: "info"})
	outInfo, stderr, err := utils.ExecCommand(pfsInfoCmd, 20*time.Second)
	if err != nil {
		smslog.Debugf("pfs info %s failed, stdout: %s, stderr: %s, err: %s", deviceName, outInfo, stderr, err)
		return 0, fmt.Errorf("failed exec command %s err %s", pfsInfoCmd, err)
	}
	smslog.Infof("Successfully find block volume %s info, result [%s]", pbdName, outInfo)

	columns := strings.Split(outInfo, DefaultLineSplitStr)
	if len(columns) < 2 {
		return 0, fmt.Errorf("could not find pfs info %s is %v", pbdName, columns)
	}

	blkTagInfo := strings.Split(columns[1], DefaultRowSpaceSplitStr)
	if len(blkTagInfo) < 9 {
		return 0, fmt.Errorf("could not find pfs blktag info %s is %v", columns[1], blkTagInfo)
	}

	var nallSize int
	nallSize, err = strconv.Atoi(strings.Trim(blkTagInfo[8], ","))
	if err != nil {
		return 0, fmt.Errorf("could not change nall size [%s] string to int: %v", blkTagInfo[8], err)
	}
	smslog.Infof("successfully exec pfs info %s", deviceName)

	return int64(nallSize * PfsNallNum * exec.MiB), nil
}

func (p *Pfs) FormatFilesystem(deviceName string) error {
	/*
		##[PFS_LOG] May 14 14:03:50.202640 INF [21514] open device cluster disk, devname mapper_xhuqvcg-polar-rwo-g55236264rg-16016, flags 0x13
		##[PFS_LOG] May 14 14:03:50.203095 INF [21514] disk dev path: /dev/mapper/xhuqvcg-polar-rwo-g55236264rg-16016
		##[PFS_LOG] May 14 14:03:50.203108 INF [21514] open local disk: open(/dev/mapper/xhuqvcg-polar-rwo-g55236264rg-16016, 0x4002)
		##[PFS_LOG] May 14 14:03:50.203512 INF [21514] ioctl status 0
		##[PFS_LOG] May 14 14:03:50.203523 INF [21514] pfs_diskdev_info get pi_pbdno 0, pi_rwtype 1, pi_unitsize 4194304, pi_chunksize 10737418240, pi_disksize 42949672960
		##[PFS_LOG] May 14 14:03:50.203535 INF [21514] pfs_diskdev_info waste size: 0
		##[PFS_LOG] May 14 14:03:50.203545 INF [21514] disk size 0xa00000000, chunk size 0x280000000
		##[PFS_LOG] May 14 14:03:50.203918 INF [21514] mkfs runs forcedly, although PBD mapper_xhuqvcg-polar-rwo-g55236264rg-16016 chunk 0 is already formatted
		##Init chunk 0
		##		metaset        0/1: sectbda           0x1000, npage       80, objsize  128, nobj 2560, oid range [       0,      a00)
		##		metaset        0/2: sectbda          0x51000, npage       64, objsize  128, nobj 2048, oid range [       0,      800)
		##		metaset        0/3: sectbda          0x91000, npage       64, objsize  128, nobj 2048, oid range [       0,      800)
		##
		##Init chunk 1
		##		metaset        1/1: sectbda      0x280001000, npage       80, objsize  128, nobj 2560, oid range [    1000,     1a00)
		##		metaset        1/2: sectbda      0x280051000, npage       64, objsize  128, nobj 2048, oid range [     800,     1000)
		##		metaset        1/3: sectbda      0x280091000, npage       64, objsize  128, nobj 2048, oid range [     800,     1000)
		##
		##Init chunk 2
		##		metaset        2/1: sectbda      0x500001000, npage       80, objsize  128, nobj 2560, oid range [    2000,     2a00)
		##		metaset        2/2: sectbda      0x500051000, npage       64, objsize  128, nobj 2048, oid range [    1000,     1800)
		##		metaset        2/3: sectbda      0x500091000, npage       64, objsize  128, nobj 2048, oid range [    1000,     1800)
		##
		##Init chunk 3
		##		metaset        3/1: sectbda      0x780001000, npage       80, objsize  128, nobj 2560, oid range [    3000,     3a00)
		##		metaset        3/2: sectbda      0x780051000, npage       64, objsize  128, nobj 2048, oid range [    1800,     2000)
		##		metaset        3/3: sectbda      0x780091000, npage       64, objsize  128, nobj 2048, oid range [    1800,     2000)
		##
		##Inited filesystem(42949672960 bytes), 4 chunks, 2560 blktags, 2048 direntries, 2048 inodes per chunk
		##making paxos file
		##init paxos lease
		##making journal file
		##pfs mkfs succeeds!
	*/
	smslog.Infof("VolumeInfo %s appears to be unformatted, attempting to format as type: pfs", deviceName)
	pbdName, err := common.GetPBDName(deviceName)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed get DevicePath by deviceName %s", deviceName))
	}

	pfsMkfsCmd := getPfsCmd(pbdName, PfsOptions{command: "mkfs"})
	blockDevBytes, err := dmhelper.GetBlockDevSize(deviceName)
	if err != nil {
		return err
	}
	reqSizeIn100GiB := blockDevBytes / (100 * 1024 * 1024 * 1024)
	stdout, stderr, err := utils.ExecCommand(pfsMkfsCmd, time.Duration(20+10*reqSizeIn100GiB)*time.Second)
	if err != nil {
		smslog.Debugf("pfs mkfs %s failed, stdout: %s, stderr: %s, err: %s", deviceName, stdout, stderr, err)
		return fmt.Errorf("failed exec command %s err %s", pfsMkfsCmd, err)
	}
	smslog.Infof("VolumeInfo successfully formatted (mkfs): pfs - %s", deviceName)

	return nil
}

func (p *Pfs) ExpandFilesystem(deviceName string, expandCapacity int64, originCapacity int64) error {
	/*
		# pfs 扩容
		pfs -C disk growfs -o oldChunkNum -n newChunkNum mapper_${volumeName}
		## e.g., output
		## >pfs -C disk growfs -o 1 -n 3 mapper_xhuqv2l-online-expand-pvc-01
		## Init chunk 1
		## metaset        1/1: sectbda      0x280001000, npage       80, objsize  128, nobj 2560, oid range [    1000,     1a00)
		## metaset        1/2: sectbda      0x280051000, npage       64, objsize  128, nobj 2048, oid range [     800,     1000)
		## metaset        1/3: sectbda      0x280091000, npage       64, objsize  128, nobj 2048, oid range [     800,     1000)
		##
		## Init chunk 2
		## metaset        2/1: sectbda      0x500001000, npage       80, objsize  128, nobj 2560, oid range [    2000,     2a00)
		## metaset        2/2: sectbda      0x500051000, npage       64, objsize  128, nobj 2048, oid range [    1000,     1800)
		## metaset        2/3: sectbda      0x500091000, npage       64, objsize  128, nobj 2048, oid range [    1000,     1800)
	*/
	smslog.Infof("VolumeInfo %s appears to be expanded", deviceName)
	blockDevBytes, err := dmhelper.GetBlockDevSize(deviceName)
	if err != nil {
		return err
	}
	deviceBlockGiB := exec.BytesToGiB(blockDevBytes)

	expandCapacityGiB := exec.BytesToGiB(expandCapacity)
	if deviceBlockGiB < expandCapacityGiB {
		return fmt.Errorf("please check rescan device and multipathd resize map, block device capacity(%dGiB) not equals request capacity(%dGiB)", deviceBlockGiB, expandCapacityGiB)
	}

	originGiB := exec.BytesToGiB(originCapacity)
	oldChunkNum := int(originGiB / PfsChunkNum)

	newChunkNum := int(expandCapacityGiB / PfsChunkNum)
	if oldChunkNum == newChunkNum {
		smslog.Infof("Successfully expanded block volume, pfs oldChunkNum [%d] is equals to newChunkNum [%d]", oldChunkNum, newChunkNum)
		return nil
	}

	pbdName, err := common.GetPBDName(deviceName)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed get PBDName by deviceName %s", deviceName))
	}
	pfsExpandCmd := getPfsCmd(pbdName, PfsOptions{command: "growfs", oldChunkNum: oldChunkNum, newChunkNum: newChunkNum})
	reqSizeIn100GiB := blockDevBytes / (100 * 1024 * 1024 * 1024)
	stdout, stderr, err := utils.ExecCommand(pfsExpandCmd, time.Duration(20+10*reqSizeIn100GiB)*time.Second)
	if err != nil {
		smslog.Debugf("pfs growfs %s failed, stdout: %s, stderr: %s, err: %s", deviceName, stdout, stderr, err)
		return fmt.Errorf("failed exec command %s err %s", pfsExpandCmd, err)
	}
	smslog.Infof("successfully exec pfs growfs %s", deviceName)

	return nil
}

func getPfsCmd(pbdName string, options PfsOptions) string {
	pfsPrefix := fmt.Sprintf("pfs -C disk")
	/*
		# cmd = []string{"pfs", -C", "disk", "info", fmt.Sprintf("mapper_" + volumeID)}
		# cmd = []string{"pfs", -C", "disk", "mkfs", "-u", "30", fmt.Sprintf("mapper_" + volumeID)}
		# cmd = []string{"pfs", -C", "disk", "growfs", "-o", oldChunkNum, "-n", newChunkNum, fmt.Sprintf("mapper_" + volumeID)}
	*/
	var midfix string
	switch options.command {
	case "info":
		midfix = options.command
	case "mkfs":
		midfix = fmt.Sprintf("mkfs -u 30 -l 1073741824 -f")
	case "growfs":
		midfix = fmt.Sprintf("growfs -f -o %d -n %d", options.oldChunkNum, options.newChunkNum)
	}

	return fmt.Sprintf("%s %s %s", pfsPrefix, midfix, pbdName)
}

func NewPfs() Filesystem {
	return &Pfs{}
}
