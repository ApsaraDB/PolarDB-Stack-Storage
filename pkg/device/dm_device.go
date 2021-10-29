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

package device

import (
	"fmt"
	"math"
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
	"strconv"
	"strings"
)

/**
system's page_size: getconf PAGE_SIZE

DM TABLE FORMAT
       Each line of the table specifies a single target and is of the form:

       logical_start_sector num_sectors target_type target_args

       Simple target types and target args include:

       linear destination_device start_sector
              The traditional linear mapping.

       striped num_stripes chunk_size [destination start_sector]...
              Creates a striped area.
              e.g. striped 2 32 /dev/hda1 0 /dev/hdb1 0 will map the first
              chunk (16k) as follows:
                      LV chunk 1-> hda1, chunk 1
                      LV chunk 2-> hdb1, chunk 1
                      LV chunk 3-> hda1, chunk 2
                      LV chunk 4-> hdb1, chunk 2
                      etc.

       error  Errors any I/O that goes to this area.  Useful for testing or
              for creating devices with holes in them.

       zero   Returns blocks of zeroes on reads.  Any data written is
              discarded silently.  This is a block-device equivalent of the
              /dev/zero character-device data sink described in null(4).

       More complex targets include:

       cache  Improves performance of a block device (eg, a spindle) by
              dynamically migrating some of its data to a faster smaller
              device (eg, an SSD).

       crypt  Transparent encryption of block devices using the kernel
              crypto API.

       delay  Delays reads and/or writes to different devices.  Useful for
              testing.

       flakey Creates a similar mapping to the linear target but exhibits
              unreliable behaviour periodically.  Useful for simulating
              failing devices when testing.

       mirror Mirrors data across two or more devices.

       multipath
              Mediates access through multiple paths to the same device.

       raid   Offers an interface to the kernel's software raid driver, md.

       snapshot
              Supports snapshots of devices.

       thin, thin-pool
              Supports thin provisioning of devices and also provides a
              better snapshot support.

       To find out more about the various targets and their table formats
       and status lines, please read the files in the Documentation/device-
       mapper directory in the kernel source tree.  (Your distribution might
       include a copy of this information in the documentation directory for
       the device-mapper package.)
EXAMPLES
       # A table to join two disks together
       0 1028160 linear /dev/hda 0
       1028160 3903762 linear /dev/hdb 0
       # A table to stripe across the two disks,
       # and add the spare space from
       # hdb to the back of the volume
       0 2056320 striped 2 32 /dev/hda 0 /dev/hdb 0
       2056320 2875602 linear /dev/hdb 1028160
*/

const (
	MinimalLen              int   = 3
	StripedSize             int   = 2
	StripedChunkSizeInBytes int   = 128 * 1024
	DefaultOffsetSector     int64 = 8192
)

const (
	NewLineSign   = "\n"
	BlankSign     = " "
	EqualSign     = "="
	SemicolonSign = ";"
	CommaSign     = ","
)

const (
	UnknownType DmDeviceType = "unknown"
	Linear      DmDeviceType = "linear"
	Striped     DmDeviceType = "striped"
	Mirror      DmDeviceType = "mirror"
	Multipath   DmDeviceType = "multipath"
)

const (
	LogicalStartSectorOffset = 0
	NumSectorsOffset         = 1
	TargetTypeOffset         = 2
	TargetArgsOffset         = 3
)

const (
	WwidKey         string = "wwid"
	VendorKey       string = "vendor"
	PathsKey        string = "paths"
	PathNumKey      string = "pathNum"
	ExtendKey       string = "extend"
	DmTableItemKey  string = "dmTableItem"
	DmTableItemsKey string = "dmTableItems"
)

type DmTarget interface {
	SetValue(key string, value interface{})
	GetValue(key string) (interface{}, bool)
	String() string
}

type DmChild struct {
	ChildType  common.LvType `json:"child_type"`
	ChildId    string        `json:"child_id"`
	SectorSize int           `json:"sector_size"`
	Sectors    int64         `json:"sectors"`
}

type DmDeviceCore struct {
	VolumeId   string       `json:"volume_id"`
	DeviceType DmDeviceType `json:"device_type"`
	SectorNum  int64        `json:"sector_num"`
	SectorSize int          `json:"sector_size"`
	Children   []*DmChild   `json:"children"`
}

func (d *DmDeviceCore) GetDmTableString() (string, error) {
	if d.DeviceType == Linear {
		dmDevice, err := ParseLinearDevice(d)
		if err != nil {
			return "", err
		}
		return dmDevice.String(), nil
	}
	if d.DeviceType == Striped {
		dmDevice, err := ParseStripedDevice(d)
		if err != nil {
			return "", err
		}
		return dmDevice.String(), nil
	}
	return "", fmt.Errorf("unsupport dm type %s", d.DeviceType)
}

type DmDeviceType string
type DmDevice struct {
	Name            string           `json:"name"`
	VolumeId        string           `json:"volume_id"`
	DeviceType      DmDeviceType     `json:"device_type"`
	SectorNum       int64            `json:"sector_num"`
	SectorSize      int              `json:"sector_size"`
	FsType          common.FsType    `json:"fs_type"`
	FsSize          int64            `json:"fs_size"`
	PrSupportStatus *PrSupportReport `json:"pr_support_status"`
	UsedSize        int64            `json:"used_size"`
	SerialNumber    string           `json:"serial_number"`
	DmTarget
}

func (d *DmDevice) Id() string {
	return d.Name
}

func (d *DmDevice) String() string {
	return d.DmTarget.String()
}

/*
	校验target的边界是否合法，是否存在空洞，是否重叠
	标准合法的应该是最小和最大值只有一个，中间各个边界点有两个
*/
//todo fine tune the ugly code
func (d *DmDevice) Validate() bool {
	var (
		max  int64 = -1
		min  int64 = math.MaxInt64
		cnts       = make(map[int64]int)
	)

	addHandler := func(k int64, cnts map[int64]int) error {
		v, ok := cnts[k]
		if !ok {
			v = 1
		} else {
			v += 1
		}

		if v > 2 {
			return fmt.Errorf("item bound %d is illegal", k)
		}

		cnts[k] = v
		return nil
	}

	dmTableItems, ok := d.DmTarget.GetValue(DmTableItemsKey)
	if !ok {
		smslog.Infof("can not find dmtableitems from %v", d.DmTarget)
		return true
	}
	for _, item := range dmTableItems.([]*DmTableItem) {
		t := item.TargetArgs.(LinearArgs)
		start := t.StartSector
		end := t.StartSector + int64(item.NumSectors)

		if start < 0 || end < 0 || start >= end {
			smslog.Errorf("invalid item: %+v", item)
			return false
		}

		if min > start {
			min = start
		}

		if max < end {
			max = end
		}

		if err := addHandler(start, cnts); err != nil {
			smslog.Errorf("invalid table, err: %s", err)
			return false
		}

		if err := addHandler(end, cnts); err != nil {
			smslog.Errorf("invalid table, err: %s", err)
			return false
		}
	}

	for k, v := range cnts {
		if k == max || k == min {
			if v != 1 {
				smslog.Errorf("invalid item bound: %d", k)
				return false
			}
		} else {
			if v != 2 {
				smslog.Errorf("invalid item bound: %d", k)
				return false
			}
		}
	}

	return true
}

func (d *DmDevice) Compare(other *DmDevice) bool {
	if d.Id() == other.Id() {
		return true
	}
	return false
}

func (d *DmDevice) Children() []string {
	switch d.DeviceType {
	case Multipath:
		return []string{d.VolumeId}
	case Linear:
		return d.DmTarget.(*LinearDeviceTarget).GetChildren()
	case Striped:
		return d.DmTarget.(*StripedDeviceTarget).GetChildren()
	}
	return []string{}
}

type DmTableItem struct {
	LogicalStartSector int64
	NumSectors         int64
	TargetArgs         interface{}
}

type MultipathDeviceTarget struct {
	Name            string
	Product         string
	Wwid            string
	Vendor          string
	Paths           []string
	PathNum         int
	PrSupportStatus string
	Extend          string
	DmTableItem     *DmTableItem
}

func (t *MultipathDeviceTarget) SetValue(key string, value interface{}) {
	switch key {
	case WwidKey:
		t.Wwid = value.(string)
	case VendorKey:
		t.Vendor = value.(string)
	case PathNumKey:
		t.PathNum = value.(int)
	case PathsKey:
		t.Paths = value.([]string)
	case ExtendKey:
		t.Extend = value.(string)
	case DmTableItemKey:
		t.DmTableItem = value.(*DmTableItem)
	}
}

func (t *MultipathDeviceTarget) GetValue(key string) (interface{}, bool) {
	switch key {
	case WwidKey:
		return t.Wwid, true
	case VendorKey:
		return t.Vendor, true
	case PathNumKey:
		return t.PathNum, true
	case PathsKey:
		return t.Paths, true
	case ExtendKey:
		return t.Extend, true
	case DmTableItemKey:
		return t.DmTableItem, true
	default:
		return nil, false
	}
}

func (t *MultipathDeviceTarget) String() string {
	return ""
}

type LinearDeviceTarget struct {
	DmTableItems []*DmTableItem
}

func (t *LinearDeviceTarget) SetValue(key string, value interface{}) {
	switch key {
	case DmTableItemsKey:
		t.DmTableItems = value.([]*DmTableItem)
	}
}

func (t *LinearDeviceTarget) GetValue(key string) (interface{}, bool) {
	switch key {
	case DmTableItemsKey:
		return t.DmTableItems, true
	default:
		return nil, false
	}
}

func (t *LinearDeviceTarget) String() string {
	lines := make([]string, 0)
	for _, item := range t.DmTableItems {
		linearArgs := item.TargetArgs.(*LinearArgs)
		line := strings.Join(
			[]string{
				fmt.Sprintf("%d %d %s", item.LogicalStartSector, item.NumSectors, string(Linear)),
				linearArgs.String(),
			},
			BlankSign)
		lines = append(lines, line)
	}
	return strings.Join(lines, NewLineSign)
}

func (t *LinearDeviceTarget) GetChildren() []string {
	ret := make([]string, 0)
	for _, item := range t.DmTableItems {
		child := item.TargetArgs.(*LinearArgs).TargetDevice
		if child != nil {
			ret = append(ret, child.Name)
		}
	}
	return ret
}

type StripedDeviceTarget struct {
	DmTableItems []*DmTableItem
}

func (t *StripedDeviceTarget) SetValue(key string, value interface{}) {
	switch key {
	case DmTableItemsKey:
		t.DmTableItems = value.([]*DmTableItem)
	}
}

func (t *StripedDeviceTarget) GetValue(key string) (interface{}, bool) {
	switch key {
	case DmTableItemsKey:
		return t.DmTableItems, true
	default:
		return nil, false
	}
}

func (t *StripedDeviceTarget) String() string {
	if len(t.DmTableItems) == 0 {
		return ""
	}
	item0 := t.DmTableItems[0]
	stripArgs := item0.TargetArgs.(*StripedArgs)
	return fmt.Sprintf("%d %d %s %s",
		item0.LogicalStartSector,
		item0.NumSectors,
		string(Striped),
		stripArgs.String())
}

func (t *StripedDeviceTarget) GetChildren() []string {
	ret := make([]string, 0)
	for _, item := range t.DmTableItems {
		children := item.TargetArgs.(*StripedArgs).TargetList
		for _, child := range children {
			childDevice := child.TargetDevice
			if childDevice != nil {
				ret = append(ret, childDevice.Name)
			}
		}
	}
	return ret
}

func NewLinearDmItem(start int64, numSectors int64, targetDevice string, offsetSector int64) *DmTableItem {
	return &DmTableItem{
		LogicalStartSector: start,
		NumSectors:         numSectors,
		TargetArgs: &LinearArgs{&BaseArgs{
			TargetDevice: &DmDevice{
				Name: targetDevice,
			},
			StartSector: offsetSector,
		}},
	}
}

func NewStripedDmItem(start int64, numSectors int64, sectorSize int, targetDevices []string) *DmTableItem {
	stripedArgs := StripedArgs{
		NumStripes: len(targetDevices),
		ChunkSize:  int64(StripedChunkSizeInBytes / sectorSize),
	}
	for _, tgtDevice := range targetDevices {
		stripedArgs.TargetList = append(stripedArgs.TargetList, BaseArgs{
			TargetDevice: &DmDevice{
				Name: tgtDevice,
			},
			StartSector: 0,
		})
	}
	return &DmTableItem{
		LogicalStartSector: start,
		NumSectors:         numSectors,
		TargetArgs:         &stripedArgs,
	}
}

//RefDevice VolumeId
type BaseArgs struct {
	TargetDevice *DmDevice
	StartSector  int64
}

func (a *BaseArgs) String() string {
	return fmt.Sprintf("/dev/mapper/%s %d", a.TargetDevice.Name, a.StartSector)
}

type LinearArgs struct {
	*BaseArgs
}

type MultipathArgs struct {
	PathNum int
	Paths   []string
}

type StripedArgs struct {
	NumStripes int
	ChunkSize  int64
	TargetList []BaseArgs
}

func (a *StripedArgs) String() string {
	argStrs := []string{
		fmt.Sprintf("%d", a.NumStripes),
		fmt.Sprintf("%d", a.ChunkSize),
	}

	for _, baseArg := range a.TargetList {
		argStrs = append(argStrs, baseArg.String())
	}

	return strings.Join(argStrs, " ")
}

func ParseFromLine(line string) (*DmTableItem, *DmDeviceType, error) {
	parts := strings.Split(strings.TrimSpace(line), BlankSign)
	if len(parts) < MinimalLen {
		err := fmt.Errorf("wrong format for line %s, the items in line is less than %d", line, MinimalLen)
		smslog.Infof("%s", err.Error())
		return nil, nil, err
	}
	item := &DmTableItem{}

	var startIdx = 0
	if strings.Contains(parts[startIdx], ":") {
		startIdx = 1
	}
	lss, err := strconv.ParseInt(parts[startIdx+LogicalStartSectorOffset], 10, 0)
	if err != nil {
		smslog.Errorf("err when parse line %s, %v", line, err)
		return nil, nil, err
	}
	item.LogicalStartSector = lss

	ns, err := strconv.ParseInt(parts[startIdx+NumSectorsOffset], 10, 0)
	if err != nil {
		smslog.Errorf("err when parse line %s, %v", line, err)
		return nil, nil, err
	}
	item.NumSectors = ns

	tt := DmDeviceType(parts[startIdx+TargetTypeOffset])
	args, err := parseTargetArgs(tt, parts[startIdx+TargetArgsOffset:])
	if err != nil {
		return nil, nil, err
	}
	item.TargetArgs = args
	return item, &tt, nil
}

func parseTargetArgs(t DmDeviceType, argStrs []string) (interface{}, error) {
	switch t {
	case Linear:
		return parseLinearArgs(argStrs)
	case Multipath:
		return parseMultipathArgs(argStrs)
	case Striped:
		return parseStripedArgs(argStrs)
	default:
		return nil, fmt.Errorf("still not support the type: (%s)", t)
	}
}

func parseBaseArgs(args []string) (*BaseArgs, error) {
	ret := &BaseArgs{}
	if len(args) != 2 {
		err := fmt.Errorf("wrong linear arg format %v", args)
		smslog.Infof("%s", err.Error())
		return nil, err
	}

	ret.TargetDevice = &DmDevice{
		Name: args[0],
	}

	ss, err := strconv.ParseInt(args[1], 10, 0)
	if err != nil {
		smslog.Errorf("Parse LinearArgs StartSector %s err %v", args, err)
		return nil, err
	}
	ret.StartSector = ss
	return ret, nil
}

func parseLinearArgs(parts []string) (interface{}, error) {
	ret := &LinearArgs{}
	baseArgs, err := parseBaseArgs(parts)
	if err != nil {
		return nil, err
	}
	ret.BaseArgs = baseArgs
	return ret, nil
}

func parseStripedArgs(parts []string) (interface{}, error) {
	ret := &StripedArgs{}

	//try return error last
	if len(parts) < 3 {
		err := fmt.Errorf("wrong striped args format %v", parts)
		smslog.Infof("%v", err)
		return nil, err
	}

	ns, err := strconv.ParseInt(parts[0], 10, 0)
	if err != nil {
		smslog.Errorf("Parse StripedArgs numberStripes err %s, %v", parts[0], err)
		return nil, err
	}
	ret.NumStripes = int(ns)

	cs, err := strconv.ParseInt(parts[1], 10, 0)
	if err != nil {
		smslog.Errorf("Parse StripedArgs chunkSize err %s, %v", parts[1], err)
		return nil, err
	}
	ret.ChunkSize = cs

	if (len(parts) - 2) != 2*ret.NumStripes {
		err := fmt.Errorf("StripArgs parse err for devices nuber wrong %v", parts)
		smslog.Info(err.Error())
		return nil, err
	}

	for i := 0; i < ret.NumStripes; i++ {
		startIdx := 2 + 2*i
		baseArgs, err := parseBaseArgs(parts[startIdx : startIdx+2])
		if err != nil {
			return nil, err
		}
		ret.TargetList = append(ret.TargetList, *baseArgs)
	}
	return ret, nil
}

//todo fix this
func parseMultipathArgs(parts []string) (interface{}, error) {
	ret := &MultipathArgs{}
	return ret, nil
}

type PrSupportReport struct {
	Pr7Supported bool   `json:"pr7Supported"`
	PrCapacities string `json:"prCapacities"`
	Extend       string `json:"extend"`
	PrKey        string `json:"pr_node_id"`
}

func (p *PrSupportReport) String() string {
	bytes, err := common.StructToBytes(p)
	if err != nil {
		return err.Error()
	}
	return string(bytes)
}

func PrSupportReportFromString(str string) *PrSupportReport {
	ret := &PrSupportReport{}
	if str == "" {
		return ret
	}
	err := common.BytesToStruct([]byte(str), ret)
	if err != nil {
		smslog.Errorf("PrSupportReportFromString err %s ", err.Error())
	}
	return ret
}

func ParseLinearDevice(deviceCore *DmDeviceCore) (*DmDevice, error) {
	retDevice := &DmDevice{
		DeviceType: deviceCore.DeviceType,
		DmTarget: &LinearDeviceTarget{
			DmTableItems: make([]*DmTableItem, 0),
		},
	}
	var (
		totalSectorNum int64
		sectorSize     int
		dmTableItems   = make([]*DmTableItem, 0)
	)
	for _, child := range deviceCore.Children {
		numSectors := child.Sectors - DefaultOffsetSector
		dmTableItems = append(dmTableItems, NewLinearDmItem(totalSectorNum, numSectors, child.ChildId, DefaultOffsetSector))
		totalSectorNum += numSectors
		if sectorSize != 0 && sectorSize != child.SectorSize {
			return nil, fmt.Errorf("sector size not equal %d, %d", sectorSize, child.SectorSize)
		}
		sectorSize = child.SectorSize
	}
	retDevice.SectorNum = totalSectorNum
	retDevice.SectorSize = sectorSize
	retDevice.DmTarget.SetValue(DmTableItemsKey, dmTableItems)
	return retDevice, nil
}

func ParseStripedDevice(deviceCore *DmDeviceCore) (*DmDevice, error) {
	stripedDevice := &DmDevice{
		DeviceType: Striped,
		DmTarget: &StripedDeviceTarget{
			DmTableItems: make([]*DmTableItem, 0),
		},
	}
	var (
		totalSectorNum int64
		sectorSize     int
		sectorNum      int64
		dmTableItems   = make([]*DmTableItem, 0)
	)

	paths := make([]string, 0)
	for _, child := range deviceCore.Children {
		sectorNum = 0
		if sectorSize != 0 && sectorSize != child.SectorSize {
			return nil, fmt.Errorf("sector size not equal device 1 [%d], device 2 [%d]", sectorSize, child.SectorSize)
		}
		sectorNum = child.Sectors
		sectorSize = child.SectorSize
		paths = append(paths, child.ChildId)
		totalSectorNum += sectorNum
	}
	dmTableItems = append(dmTableItems, NewStripedDmItem(0, totalSectorNum, sectorSize, paths))
	stripedDevice.SectorNum = totalSectorNum
	stripedDevice.SectorSize = sectorSize
	stripedDevice.DmTarget.SetValue(DmTableItemsKey, dmTableItems)
	return stripedDevice, nil
}
