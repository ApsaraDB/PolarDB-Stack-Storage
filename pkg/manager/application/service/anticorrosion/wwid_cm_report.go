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

package anticorrosion

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/config"
	"polardb-sms/pkg/protocol"
	"time"
)

const (
	KubeSystemNamespace = "kube-system"
	WwidCmPrefix        = "cloud-provider-wwid-usage-"
	cmLabelKey          = "aliyun.polar.box.wwid.ccm.watch"
	cmLabelValue        = "node.daemon.set"
)

type MultiPathObj struct {
	WWID    string `json:"wwid"`
	Alias   string `json:"alias,omitempty"`
	SizeMB  int64  `json:"size_mb"`
	Vendor  string `json:"vendor,omitempty"`
	WP      string `json:"wp,omitempty"`
	Product string `json:"product,omitempty"`
	VenProd string `json:"ven_prod,omitempty"`
	Device  string `json:"device,omitempty"`
	PathNum int    `json:"path_num,omitempty"`

	Major      string `json:"major,omitempty"`
	Minor      string `json:"minor,omitempty"`
	SectorNum  int64  `json:"sector_num,omitempty"`
	SectorSize int64  `json:"sector_size,omitempty"`
	BlockSize  int64  `json:"block_size,omitempty"`

	CreateTime    time.Time `json:"create_time,omitempty"`
	LastCheckTime time.Time `json:"last_check_time,omitempty"`
	DeleteTime    time.Time `json:"delete_time,omitempty"`

	IsDeleted bool `json:"is_deleted,omitempty"`
}

func getWwidCmByNodeId(nodeId string) (*corev1.ConfigMap, error) {
	var (
		cm     *corev1.ConfigMap
		err    error
		cmName = fmt.Sprintf("%s%s", WwidCmPrefix, nodeId)
	)
	err = common.RunWithRetry(3, 1*time.Second, func(retryTimes int) error {
		cm, err = config.ClientSet.CoreV1().
			ConfigMaps(KubeSystemNamespace).
			Get(cmName, metav1.GetOptions{})
		return err
	})
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func createWwidCmByNodeId(nodeId string) (*corev1.ConfigMap, error) {
	var (
		cm     *corev1.ConfigMap
		err    error
		cmName = fmt.Sprintf("%s%s", WwidCmPrefix, nodeId)
	)
	cm = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: KubeSystemNamespace,
			Labels: map[string]string{
				cmLabelKey: cmLabelValue,
			},
		},
		Data: map[string]string{},
	}
	err = common.RunWithRetry(3, 1*time.Second, func(retryTimes int) error {
		_, err = config.ClientSet.CoreV1().
			ConfigMaps(KubeSystemNamespace).
			Create(cm)
		return err
	})
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func BatchUpdateByEvents(batchEvent *protocol.BatchEvent) {
	cm, err := getWwidCmByNodeId(batchEvent.NodeId)
	if err != nil {
		if errors.IsNotFound(err) {
			cm, err = createWwidCmByNodeId(batchEvent.NodeId)
			if err != nil {
				smslog.Infof("BatchUpdateByEvents: createWwidCmByNodeId err %s", err.Error())
				return
			}
		} else {
			smslog.Infof("BatchUpdateByEvents: getWwidCmByNodeId err %s", err.Error())
			return
		}
	}

	var hasChanged = false
	var createTime = time.Now()
	var checkTime = time.Now()
	for _, event := range batchEvent.Events {
		lun := protocol.Lun{}
		if err := protocol.Decode(event.Body, &lun); err != nil {
			smslog.Errorf("LunAddEvent: could not decode event %s: %s", event.Body, err.Error())
			return
		}
		if lun.VolumeId == "" || lun.Name == "" {
			continue
		}
		currentLunStr, exist := cm.Data[lun.VolumeId]
		if !exist {
			//create
			newLun := &MultiPathObj{
				Alias:         lun.Name,
				WWID:          lun.VolumeId,
				Vendor:        lun.Vendor,
				Product:       lun.Product,
				VenProd:       fmt.Sprintf("%s %s", lun.Vendor, lun.Product),
				PathNum:       lun.PathNum,
				SectorNum:     lun.Sectors,
				SectorSize:    int64(lun.SectorSize),
				BlockSize:     lun.Size,
				SizeMB:        lun.Size / 1024 / 1024,
				CreateTime:    createTime,
				LastCheckTime: checkTime,
				IsDeleted:     false,
			}
			newLunBytes, err := common.StructToBytes(newLun)
			if err != nil {
				smslog.Infof("BatchUpdateByEvents newLun to bytes err %s", err.Error())
				continue
			}
			cm.Data[lun.VolumeId] = string(newLunBytes)
			hasChanged = true
			smslog.Debugf("create new device id [%s], name [%s]", lun.VolumeId, lun.Name)
		} else {
			//update
			currentLun := &MultiPathObj{}
			err = common.BytesToStruct([]byte(currentLunStr), currentLun)
			if err != nil {
				smslog.Infof("BatchUpdateByEvents update lun err %s", err.Error())
				continue
			}
			currentLun.LastCheckTime = checkTime
			currentLun.PathNum = lun.PathNum
			currentLun.SectorSize = int64(lun.SectorSize)
			currentLun.SectorNum = lun.Sectors
			currentLun.Alias = lun.Name
			currentLun.BlockSize = lun.Size
			currentLun.SizeMB = lun.Size / 1024 / 1024
			if currentLun.IsDeleted {
				currentLun.IsDeleted = false
				currentLun.DeleteTime = time.Time{}
			}
			updateLunBytes, err := common.StructToBytes(currentLun)
			if err != nil {
				smslog.Infof("BatchUpdateByEvents updateLun to bytes err %s", err.Error())
				continue
			}
			cm.Data[lun.VolumeId] = string(updateLunBytes)
			hasChanged = true
		}
	}

	for _, lunStr := range cm.Data {
		currentLun := &MultiPathObj{}
		err = common.BytesToStruct([]byte(lunStr), currentLun)
		if err != nil {
			smslog.Infof("BatchUpdateByEvents update lun err %s", err.Error())
			continue
		}
		if !currentLun.LastCheckTime.Equal(checkTime) && !currentLun.IsDeleted {
			smslog.Debugf("delete device :name [%s] id [%s] on node %s", currentLun.Alias, currentLun.WWID, batchEvent.NodeId)
			currentLun.IsDeleted = true
			currentLun.DeleteTime = checkTime
		} else {
			continue
		}
		updateLunBytes, err := common.StructToBytes(currentLun)
		if err != nil {
			smslog.Infof("BatchUpdateByEvents updateLun to bytes err %s", err.Error())
			continue
		}
		cm.Data[currentLun.WWID] = string(updateLunBytes)
		hasChanged = true
	}
	if !hasChanged {
		return
	}
	_ = common.RunWithRetry(3, 1*time.Second, func(retryTimes int) error {
		_, err = config.ClientSet.CoreV1().
			ConfigMaps(KubeSystemNamespace).
			Update(cm)
		return err
	})
}
