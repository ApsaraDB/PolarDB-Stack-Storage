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
	"polardb-sms/pkg/common"
	"polardb-sms/pkg/device"
	smslog "polardb-sms/pkg/log"
	"regexp"
	"strings"
)

const (
	WwidKey       = "uuid"
	PathNumKey    = "paths"
	VendorKey     = "vend"
	PathMapKey    = "path_groups"
	PathKey       = "dev"
	MajorMinorKey = "dev_st"
)

type MultipathParam struct {
	Name    string   `json:"name"`
	Wwid    string   `json:"wwid"`
	PathNum int      `json:"path_num"`
	Paths   []string `json:"paths"`
	Vendor  string   `json:"vendor"`
	Product string   `json:"product"`
}

type Path struct {
	Dev  string `json:"dev"`
	DevT string `json:"dev_t"`
}

type PathGroup struct {
	Selector string `json:"selector"`
	Pri      int    `json:"pri"`
	DmStatus string `json:"dm_st"`
	Group    int    `json:"group"`
	Paths    []Path `json:"paths"`
}

type ParamMap struct {
	Name       string      `json:"name"`
	UUID       string      `json:"uuid"`
	Sysfs      string      `json:"sysfs"`
	Vendor     string      `json:"vend"`
	Product    string      `json:"prod"`
	PathNum    int         `json:"paths"`
	PathGroups []PathGroup `json:"path_groups"`
}

type RawMultipathParam struct {
	MajorVersion int      `json:"major_version"`
	MinorVersion int      `json:"minor_version"`
	ParamMap     ParamMap `json:"map"`
}

//todo refact the ugly code
func getMultipathParam(name string) (*MultipathParam, error) {
	var (
		stdout         string
		stderr         string
		err            error
		ret            = &MultipathParam{}
		subPathPattern = regexp.MustCompile(`\d:\d:\d:\d`)
		subPathPatternNVMe = regexp.MustCompile(`#:#:#:#`)
	)
	cmd := fmt.Sprintf("multipathd show map \"%s\" topology", name)

	stdout, stderr, err = utils.ExecCommand(cmd, utils.CmdDefaultTimeout)
	if err != nil && !strings.Contains(stderr, "No such device or address") {
		return nil, fmt.Errorf("read multipathd failed, stdout: %s, stderr: %s, err: %s", stdout, stderr, err)
	}
	lines := strings.Split(stdout, device.NewLineSign)
	if len(lines) < 2 {
		return nil, fmt.Errorf("multipath result is err format %s", stdout)
	}
	smslog.Debugf("getMultipathParam: device [%s] output [%s]", name, stdout)
	firstLineFields := strings.Fields(lines[0])
	lenFields := len(firstLineFields)
	if lenFields >= 4 {
		ret.Wwid = firstLineFields[lenFields-4]
		ret.Wwid = strings.TrimLeft(ret.Wwid, "(")
		ret.Wwid = strings.TrimRight(ret.Wwid, ")")
		ret.Name = getRealVolumeName(ret.Wwid)
		ret.Vendor = firstLineFields[lenFields-2]
		ret.Product = firstLineFields[lenFields-1]
		ret.Product = strings.TrimPrefix(ret.Product, device.CommaSign)
	} else if lenFields >= 1 {
		ret.Wwid = firstLineFields[0]
		ret.Name = getRealVolumeName(ret.Wwid)
		ret.Vendor = ""
		ret.Product = ""
	} else {
		smslog.Debug("%s not a valid multipth device", name)
		return nil, fmt.Errorf("%s not a valid multipth device", name)
	}

	var paths []string
	remainLines := lines[1:]
	for _, line := range remainLines {
		if strings.Contains(line, "failed") {
			smslog.Debugf("failed path detected for %s output %s", name, line)
			continue
		}
		if subPathPattern.MatchString(line) || subPathPatternNVMe.MatchString(line) {
			subPathFields := strings.Fields(line)
			if strings.HasPrefix(line, "| ") {
				paths = append(paths, subPathFields[3])
			} else {
				paths = append(paths, subPathFields[2])
			}
		}
	}
	ret.Paths = paths
	ret.PathNum = len(paths)
	if len(paths) == 0 {
		smslog.Debugf("device %s no path can write", name)
		return nil, fmt.Errorf("device %s no path can write", name)
	}
	return ret, nil
}

func getRealVolumeName(volumeId string) string {
	realVolumeName := fmt.Sprintf("pv-%s", volumeId)
	realVolumePath := fmt.Sprintf("/dev/mapper/%s", realVolumeName)
	if common.PathExists(realVolumePath) {
		return realVolumeName
	}
	return volumeId
}
