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

package lv

import (
	"fmt"
	"polardb-sms/pkg/common"
	"polardb-sms/pkg/device"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/config"
	"polardb-sms/pkg/manager/domain"
	"polardb-sms/pkg/manager/domain/pv"
	"strings"
)

type InnerVolumeTypeAndId struct {
	VolumeId   string            `json:"volume_id"`
	VolumeType common.VolumeType `json:"volume_type"`
}

type Children struct {
	Items []domain.Volume `json:"items"`
}

func (c *Children) MarshalJSON() ([]byte, error) {
	var volumes = make([]*InnerVolumeTypeAndId, 0)
	for _, item := range c.Items {
		volumes = append(volumes, &InnerVolumeTypeAndId{
			VolumeId:   item.GetVolumeId(),
			VolumeType: item.Type(),
		})
	}
	return common.StructToBytes(&volumes)
}

func (c *Children) UnmarshalJSON(b []byte) error {
	temp := ParseChildren(string(b))
	c.Items = temp
	return nil
}

func (c *Children) AddChild(child domain.Volume) {
	if c.Items == nil {
		c.Items = make([]domain.Volume, 0)
	}
	if c.HasChild(child) {
		return
	}
	c.Items = append(c.Items, child)
}

func (c *Children) HasChild(child domain.Volume) bool {
	for _, item := range c.Items {
		if item.GetVolumeId() == child.GetVolumeId() {
			return true
		}
	}
	return false
}

func ParseChildren(str string) []domain.Volume {
	temp := make([]*InnerVolumeTypeAndId, 0)
	err := common.BytesToStruct([]byte(str), &temp)
	if err != nil {
		smslog.Errorf("ParseChildren err:  BytesToStruct %s", err.Error())
		return nil
	}
	var volumes []domain.Volume
	for _, innerVolume := range temp {
		volume, err := GetVolumeByTypeAndId(innerVolume)
		if err != nil || volume == nil {
			smslog.Errorf("failed to get volume type %s, id %s", innerVolume.VolumeType, innerVolume.VolumeId)
			return nil
		}
		volumes = append(volumes, volume)
	}
	return volumes
}

func GetVolumeByTypeAndId(innerVolume *InnerVolumeTypeAndId) (domain.Volume, error) {
	if innerVolume.VolumeType == common.Lv {
		return GetLvRepository().FindByVolumeId(innerVolume.VolumeId)
	}
	ids := strings.Split(innerVolume.VolumeId, "#")
	if len(ids) != 2 {
		return nil, fmt.Errorf("wrong param for volumeId#nodeId %s", innerVolume.VolumeId)
	}
	return pv.GetPhysicalVolumeRepository().FindByVolumeIdAndNodeId(ids[0], ids[1])
}

type PrInfo map[string]*PrCheckList

func (pi PrInfo) String() string {
	if pi == nil {
		return ""
	}
	bytes, err := common.StructToBytes(pi)
	if err != nil {
		return err.Error()
	}
	return string(bytes)
}
func (pi PrInfo) GetPrCheckListByKey(key string) *PrCheckList {
	ret := pi[key]
	if ret == nil {
		pi[key] = &PrCheckList{}
	}
	return pi[key]
}

func ParsePrInfo(str string) PrInfo {
	var ret = map[string]*PrCheckList{}
	err := common.BytesToStruct([]byte(str), &ret)
	if err != nil {
		smslog.Debugf("ParsePrInfo %s err %s", str, err.Error())
	}
	return ret
}

type Extend map[string]interface{}

func (e Extend) String() string {
	if e == nil || len(e) == 0 {
		return ""
	}
	bytes, err := common.StructToBytes(e)
	if err != nil {
		return err.Error()
	}
	return string(bytes)
}

const DmDeviceKey = "DmDeviceKey"

func (e Extend) GetDmDevice() *device.DmDevice {
	return e[DmDeviceKey].(*device.DmDevice)
}

func ParseExtend(str string) Extend {
	var ret = map[string]interface{}{}
	err := common.BytesToStruct([]byte(str), &ret)
	if err != nil {
		smslog.Debugf("ParseExtend %s err %s", str, err.Error())
	}

	return ret
}

//LogicalVolumeEntity represent a virtual volume
//in cluster and can be used by db
type LogicalVolumeEntity struct {
	domain.VolumeInfo
	LvType     common.LvType       `json:"lv_type"`
	ClusterId  int                 `json:"cluster_id"`
	PrKey      string              `json:"pr_node_id"`
	Desc       string              `json:"desc"`
	Status     domain.VolumeStatus `json:"status"`
	Children   *Children           `json:"children"`
	PrInfo     PrInfo              `json:"pr_info"`
	Extend     Extend              `json:"extend"`
	NodeIds    []string            `json:"node_ids"`
	UsedByType domain.UsedByType   `json:"used_by_type"`
	UsedByName string              `json:"used_by_name"`
}

type PrCheckList struct {
	SupportPr7        bool `json:"support_pr_7"`
	SupportPrRegister bool `json:"support_pr_register"`
	SupportPrReserve  bool `json:"support_pr_reserve"`
	SupportPrPreempt  bool `json:"support_pr_attempt"`
	SupportPrRelease  bool `json:"support_pr_release"`
	SupportPrClear    bool `json:"support_pr_clear"`
}

func (e *LogicalVolumeEntity) GetPvcName() string {
	if e.UsedByType == domain.DBUsed {
		return e.UsedByName
	}
	return ""
}

func (e *LogicalVolumeEntity) GetLvName() string {
	if e.UsedByType == domain.LvUsed {
		return e.UsedByName
	}
	return ""
}

func (e *LogicalVolumeEntity) Usable() bool {
	if e.Status.StatusValue == domain.NoAction ||
		e.Status.StatusValue == domain.Success {
		if !e.IsUsed() {
			return true
		}
	}
	return false
}

func (e *LogicalVolumeEntity) IsUsed() bool {
	if e.UsedByType <= domain.Non {
		return false
	}
	return true
}

func (e *LogicalVolumeEntity) IsLvUsed() bool {
	if e.UsedByType == domain.LvUsed {
		return true
	}
	return false
}

func (e *LogicalVolumeEntity) IsDBUsed() bool {
	if e.UsedByType == domain.DBUsed {
		return true
	}
	return false
}

func (e *LogicalVolumeEntity) SetUsedBy(name string, byType domain.UsedByType) {
	e.UsedByType = byType
	e.UsedByName = name
}

func (e *LogicalVolumeEntity) SetFsType(fsType common.FsType, fsSize int64) {
	e.FsType = fsType
	e.FsSize = fsSize
}

func (e *LogicalVolumeEntity) ReleaseUsed() {
	e.UsedByType = domain.Non
	e.UsedByName = ""
}

func (e *LogicalVolumeEntity) ClearPrKey() {
	e.PrKey = ""
}

func (e *LogicalVolumeEntity) GetCanWriteNode() config.Node {
	if e.PrKey != "" {
		ret := config.GetNodeByIp(common.PrKeyToIpV4(e.PrKey))
		if ret != nil {
			return *ret
		}
	}
	for _, node := range config.GetAvailableNodes() {
		return node
	}
	return config.Node{}
}

func (e *LogicalVolumeEntity) AddNodeId(nodeId string) {
	if nodeId == "" {
		return
	}
	for _, existNodeId := range e.NodeIds {
		if existNodeId == nodeId {
			return
		}
	}
	e.NodeIds = append(e.NodeIds, nodeId)
}

func (e *LogicalVolumeEntity) AddChildByTypeAndId(childType common.VolumeType, volumeId, nodeId string) {
	if childType == common.Pv {
		childPv, err := pv.GetPhysicalVolumeRepository().FindByVolumeIdAndNodeId(volumeId, nodeId)
		if err != nil || childPv == nil {
			return
		}
		e.Children.AddChild(childPv)
	} else {
		//TODO may be not impl like this
	}
}

func (e *LogicalVolumeEntity) Type() common.VolumeType {
	return common.Lv
}

func (e *LogicalVolumeEntity) GetDmDeviceCore() (*device.DmDeviceCore, error) {
	//todo 类型校验
	dmDeviceCore := &device.DmDeviceCore{
		VolumeId:   e.VolumeId,
		DeviceType: e.getDmDeviceType(),
		SectorNum:  e.Sectors,
		SectorSize: e.SectorSize,
		Children:   make([]*device.DmChild, 0),
		//TODO
		//TableString: e,
	}
	for _, multipathVolumeInf := range e.Children.Items {
		multipathVolume, ok := multipathVolumeInf.(*LogicalVolumeEntity)
		if !ok {
			return nil, fmt.Errorf("GetDmDeviceCore: err, do not support children type non LogicalVolumeEntity %s", e.VolumeId)
		}
		dmDeviceCore.Children = append(dmDeviceCore.Children, &device.DmChild{
			ChildType:  common.MultipathVolume,
			ChildId:    multipathVolume.GetVolumeId(),
			SectorSize: multipathVolume.GetSectorSize(),
			Sectors:    multipathVolume.GetSectors(),
		})
	}
	return dmDeviceCore, nil
}

func (e *LogicalVolumeEntity) getDmDeviceType() device.DmDeviceType {
	switch e.LvType {
	case common.MultipathVolume:
		return device.Multipath
	case common.DmLinearVolume:
		return device.Linear
	case common.DmMirrorVolume:
	case common.DmStripVolume:
		return device.Striped
	}
	return device.UnknownType
}

func (e *LogicalVolumeEntity) GetChildrenString() string {
	if e.Children == nil {
		return ""
	}
	ret, err := common.StructToBytes(e.Children)
	if err != nil {
		return ""
	}
	return string(ret)
}

func (e *LogicalVolumeEntity) GetDmDevice() (*device.DmDevice, error) {
	switch e.LvType {
	case common.DmLinearVolume:
		return ParseLinearDevice(e.Children.Items)
	case common.DmStripVolume:
		return ParseStripedDevice(e.Children.Items)
	}
	return nil, fmt.Errorf("GetDmDevice not support lvtype %s", e.LvType)
}

func (e *LogicalVolumeEntity) Valid() error {
	if e.LvType == common.DmStripVolume {
		var volumeSize int64 = -1
		for _, child := range e.Children.Items {
			if volumeSize == -1 {
				volumeSize = child.GetSectors()
				continue
			}
			if volumeSize != child.GetSectors() {
				return fmt.Errorf("stripe volume need each sub volume with same size")
			}
		}
	}
	return nil
}

func ParseLinearDevice(volumes []domain.Volume) (*device.DmDevice, error) {
	linearDevice := &device.DmDevice{
		DeviceType: device.Linear,
		DmTarget: &device.LinearDeviceTarget{
			DmTableItems: make([]*device.DmTableItem, 0),
		},
	}
	var (
		totalSectorNum int64
		sectorSize     int
		dmTableItems   = make([]*device.DmTableItem, 0)
	)
	for _, volume := range volumes {
		numSectors := volume.GetSectors() - device.DefaultOffsetSector
		dmTableItems = append(dmTableItems, device.NewLinearDmItem(totalSectorNum, numSectors, volume.GetVolumeId(), device.DefaultOffsetSector))
		totalSectorNum += numSectors
		if sectorSize != 0 && sectorSize != volume.GetSectorSize() {
			return nil, fmt.Errorf("sector size not equal %d, %d", sectorSize, volume.GetSectorSize())
		}
		sectorSize = volume.GetSectorSize()
	}
	linearDevice.SectorNum = totalSectorNum
	linearDevice.SectorSize = sectorSize
	linearDevice.DmTarget.SetValue(device.DmTableItemsKey, dmTableItems)
	return linearDevice, nil
}

func ParseStripedDevice(volumes []domain.Volume) (*device.DmDevice, error) {
	stripedDevice := &device.DmDevice{
		DeviceType: device.Striped,
		DmTarget: &device.StripedDeviceTarget{
			DmTableItems: make([]*device.DmTableItem, 0),
		},
	}
	var (
		totalSectorNum int64
		sectorSize     int
		sectorNum      int64
		dmTableItems   = make([]*device.DmTableItem, 0)
	)

	paths := make([]string, 0)
	for _, volume := range volumes {
		sectorNum = 0
		if sectorSize != 0 && sectorSize != volume.GetSectorSize() {
			return nil, fmt.Errorf("sector size not equal device 1 [%d], device 2 [%d]", sectorSize, volume.GetSectorSize())
		}
		sectorNum = volume.GetSectors()
		sectorSize = volume.GetSectorSize()
		paths = append(paths, fmt.Sprintf("/dev/mapper/%s", volume.GetVolumeId))
		totalSectorNum += sectorNum
	}
	dmTableItems = append(dmTableItems, device.NewStripedDmItem(totalSectorNum, sectorNum, sectorSize, paths))
	stripedDevice.SectorNum = totalSectorNum
	stripedDevice.SectorSize = sectorSize
	stripedDevice.DmTarget.SetValue(device.DmTableItemsKey, dmTableItems)
	return stripedDevice, nil
}
