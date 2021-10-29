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

package k8spvc

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/resource"
	"polardb-sms/pkg/common"
	"polardb-sms/pkg/manager/domain"
)

type RequestVolumeMode int

const (
	Block RequestVolumeMode = iota
	FsExt4
)

type PvcRuntimeStatus string

const (
	Pending PvcRuntimeStatus = "Pending"
	Bound                    = "Bound"
	Lost                     = "Lost"
)

type PersistVolumeClaimEntity struct {
	Name               string              `json:"name"`
	PvName             string              `json:"pv_name"`
	Namespace          string              `json:"namespace"`
	RuntimeStatus      PvcRuntimeStatus    `json:"runtime_status"`
	PvcStatus          domain.VolumeStatus `json:"pvc_status"`
	DiskStatus         *VolumeMeta         `json:"disk_status"`
	ExpectedDiskStatus *VolumeMeta         `json:"expected_status"`
	StorageClassName   string              `json:"storage_class_name"`
	RelatedWorkflow    string              `json:"related_workflow"`
	DbClusterName      string              `json:"db_cluster_name"`
	CreateTime         string              `json:"create_time"`
	PvcRepo            PvcRepository       `json:"-"`
}

func (pvc *PersistVolumeClaimEntity) ConstructPvEntity() *PersistVolumeEntity {
	pvName := fmt.Sprintf("pv-%s", pvc.DiskStatus.VolumeId)
	return &PersistVolumeEntity{
		Name:             pvName,
		StorageClassName: pvc.StorageClassName,
		Request:          pvc.ExpectedDiskStatus.DeepCopy(),
		PvcName:          pvc.Name,
		PvcNamespace:     pvc.Namespace,
	}
}

func (pvc *PersistVolumeClaimEntity) GetVolumeType() common.LvType {
	return pvc.DiskStatus.VolumeType
}

func (pvc *PersistVolumeClaimEntity) GetVolumeId() string {
	return pvc.DiskStatus.VolumeId
}

func (pvc *PersistVolumeClaimEntity) SetRequestSize(reqBytes int64) {
	pvc.ExpectedDiskStatus.Size = reqBytes
}

func (pvc *PersistVolumeClaimEntity) Release() {
	//do nothing?
	_ = pvc.SetUsed("")
}

func (pvc *PersistVolumeClaimEntity) IsLocked() bool {
	if pvc.RelatedWorkflow == domain.DummyWorkflowId {
		return false
	}
	return pvc.RelatedWorkflow != ""
}

func (pvc *PersistVolumeClaimEntity) GetLockedWorkflow() string {
	return pvc.RelatedWorkflow
}

func (pvc *PersistVolumeClaimEntity) Lock(wfl string) error {
	if pvc.IsLocked() {
		return fmt.Errorf("can not lock the pvc %s, now locked by %s", pvc.Name, wfl)
	}
	pvc.RelatedWorkflow = wfl
	_, err := pvc.PvcRepo.UpdateWorkflow(pvc)
	return err
}

func (pvc *PersistVolumeClaimEntity) UnLock() error {
	pvc.RelatedWorkflow = ""
	_, err := pvc.PvcRepo.UpdateWorkflow(pvc)
	return err
}

func (pvc *PersistVolumeClaimEntity) ReadyToUse() bool {
	return pvc.RuntimeStatus == Bound
}

func (pvc *PersistVolumeClaimEntity) Usable() bool {
	return pvc.DbClusterName == "" && pvc.RuntimeStatus == Bound
}

func (pvc *PersistVolumeClaimEntity) SetUsed(dbClusterName string) error {
	pvc.DbClusterName = dbClusterName
	_, err := pvc.PvcRepo.UpdateUsedDbCluster(pvc)
	return err
}

func (pvc *PersistVolumeClaimEntity) GetPrKey() string {
	return pvc.DiskStatus.PrKey
}

func (pvc *PersistVolumeClaimEntity) SetRequestPrKey(prKey string) {
	pvc.ExpectedDiskStatus.PrKey = prKey
}

func (pvc *PersistVolumeClaimEntity) SetPrKey(prKey string) {
	pvc.DiskStatus.PrKey = prKey
}

func (pvc *PersistVolumeClaimEntity) DeepCopy() *PersistVolumeClaimEntity {
	diskReq := pvc.DiskStatus.DeepCopy()
	return &PersistVolumeClaimEntity{
		Name:             pvc.Name,
		Namespace:        pvc.Namespace,
		RuntimeStatus:    pvc.RuntimeStatus,
		DiskStatus:       &diskReq,
		StorageClassName: pvc.StorageClassName,
		RelatedWorkflow:  pvc.RelatedWorkflow,
	}
}

type PersistVolumeEntity struct {
	Name             string     `json:"name"`
	StorageClassName string     `json:"storage_class_name"`
	Request          VolumeMeta `json:"request"`
	PvcName          string     `json:"pvc_name"`
	PvcNamespace     string     `json:"pvc_namespace"`
}

func (pv *PersistVolumeEntity) GetQuantity() (resource.Quantity, error) {
	return resource.ParseQuantity(common.BytesToGiBString(pv.Request.Size))
}

//todo change Size from string to int64
type VolumeMeta struct {
	VolumeId   string            `json:"volume_id"`
	VolumeType common.LvType     `json:"volume_type"`
	VolumeName string            `json:"volume_name"`
	Size       int64             `json:"size"`
	VolumeMode RequestVolumeMode `json:"volume_mode"`
	NeedFormat bool              `json:"need_format"`
	PrKey      string            `json:"pr_node_id"`
}

func (r VolumeMeta) DeepCopy() VolumeMeta {
	return VolumeMeta{
		VolumeId:   r.VolumeId,
		VolumeType: r.VolumeType,
		VolumeName: r.VolumeName,
		Size:       r.Size,
		VolumeMode: r.VolumeMode,
		NeedFormat: r.NeedFormat,
		PrKey:      r.PrKey,
	}
}
