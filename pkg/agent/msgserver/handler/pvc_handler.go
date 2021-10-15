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


package handler

import (
	"encoding/json"
	"fmt"
	"os"
	"polardb-sms/pkg/agent/device/filesystem"
	"polardb-sms/pkg/agent/device/reservation/mpathpersist"
	"polardb-sms/pkg/agent/device/reservation/nvme"
	"polardb-sms/pkg/agent/utils"
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/network/message"
)

type PvcCreateHandler struct {
	prExecProcessor Processor
	nodeIp          string
}

func NewPvcCreateHandler(nodeIp string) ReqHandler {
	return &PvcCreateHandler{
		prExecProcessor: mpathpersist.NewPrExecWrapper(),
		nodeIp:          nodeIp,
	}
}

func cleanOldCsiMultipathConf(volumeId string) {
	sourcePath := fmt.Sprintf("/dev/mapper/%s", volumeId)
	oldVolumeName := fmt.Sprintf("pv-%s", volumeId)
	targetPath := fmt.Sprintf("/dev/mapper/%s", oldVolumeName)
	sourceExist := common.PathExists(sourcePath)
	targetExist := common.PathExists(targetPath)
	if !sourceExist && targetExist {
		//有残留
		rmConfdCmd := fmt.Sprintf("rm -f /etc/multipath/conf.d/%s.conf", oldVolumeName)
		reLoadCmd := "multipath -r"
		utils.ExecCommand(rmConfdCmd, utils.CmdDefaultTimeout)
		utils.ExecCommand(reLoadCmd, utils.CmdDefaultTimeout)
	}
}

func (h *PvcCreateHandler) Handle(msg *message.SmsMessage) *message.SmsMessage {
	var (
		err          error
		pvcCreateCmd message.PvcCreateCommand
		ackMsgId     = msg.Head.MsgId
	)

	if err = json.Unmarshal(msg.Body.Content, &pvcCreateCmd); err != nil {
		return message.FailRespMessage(message.SmsMessageHead_CMD_PVC_CREATE_RESP, ackMsgId, err.Error())
	}

	if ok := lock(pvcCreateCmd.VolumeId); !ok {
		return message.FailRespMessage(message.SmsMessageHead_CMD_FORMAT_FS_RESP, msg.Head.MsgId, "volume is locked by another processing")
	}
	defer unlock(pvcCreateCmd.VolumeId)

	cleanOldCsiMultipathConf(pvcCreateCmd.VolumeId)
	_ = ClearLunPrInfo(pvcCreateCmd.VolumeId, h.nodeIp, h.prExecProcessor)
	if pvcCreateCmd.Format {
		switch pvcCreateCmd.FsType {
		case common.Pfs:
			err = filesystem.NewPfs().FormatFilesystem(pvcCreateCmd.VolumeId)
			if err != nil {
				return message.FailRespMessage(message.SmsMessageHead_CMD_PVC_CREATE_RESP, ackMsgId, err.Error())
			}
		case common.Ext4:
			err = filesystem.NewExt4().FormatFilesystem(pvcCreateCmd.VolumeId)
			if err != nil {
				return message.FailRespMessage(message.SmsMessageHead_CMD_PVC_CREATE_RESP, ackMsgId, err.Error())
			}
		}
	}
	return message.SuccessRespMessage(message.SmsMessageHead_CMD_PVC_CREATE_RESP, ackMsgId, nil)
}

type PvcReleaseHandler struct {
	prExecProcessor Processor
	nvmePRExecProcessor Processor
	nodeIp          string
}

func NewPvcReleaseHandler(nodeIp string) ReqHandler {
	return &PvcReleaseHandler{
		prExecProcessor: mpathpersist.NewPrExecWrapper(),
		nvmePRExecProcessor: nvme.NewNvmeExecWrapper(),
		nodeIp:          nodeIp,
	}
}

func (h *PvcReleaseHandler) Handle(msg *message.SmsMessage) *message.SmsMessage {
	var (
		err           error
		ackMsgId      = msg.Head.MsgId
		pvcReleaseCmd message.PvcReleaseCommand
	)

	if err = json.Unmarshal(msg.Body.Content, &pvcReleaseCmd); err != nil {
		return message.FailRespMessage(message.SmsMessageHead_CMD_PVC_RELEASE_RESP,
			ackMsgId, err.Error())
	}

	if ok := lock(pvcReleaseCmd.VolumeId); !ok {
		return message.FailRespMessage(message.SmsMessageHead_CMD_PVC_RELEASE_RESP, msg.Head.MsgId, "volume is locked by another processing")
	}
	defer unlock(pvcReleaseCmd.VolumeId)

	cleanOldCsiMultipathConf(pvcReleaseCmd.VolumeId)
	if err = h.cleanLunPrInfo(pvcReleaseCmd.VolumeId); err != nil {
		smslog.Infof("cleanLunPrInfo err %s", err.Error())
		return message.FailRespMessage(message.SmsMessageHead_CMD_PVC_RELEASE_RESP,
			ackMsgId, err.Error())
	}
	//if err = h.cleanSoftLinkPoint(pvcReleaseCmd.Name); err != nil {
	//	smslog.Errorf("cleanSoftLinkPoint err %s", err.Error())
	//	return message.FailRespMessage(message.SmsMessageHead_CMD_PVC_RELEASE_RESP,
	//		ackMsgId, err.Error())
	//}

	return message.SuccessRespMessage(message.SmsMessageHead_CMD_PVC_RELEASE_RESP,
		ackMsgId, nil)
}

func (h *PvcReleaseHandler) cleanSoftLinkPoint(pvName string) error {
	cleanPath := fmt.Sprintf("/dev/mapper/%s", pvName)
	if common.PathExists(cleanPath) {
		//TODO 这里如果返回err会有错误， 待排查
		return os.Remove(cleanPath)
	}
	return nil
}

func (h *PvcReleaseHandler) cleanLunPrInfo(volumeId string) error {
	if h.nvmePRExecProcessor != nil && utils.CheckNvmeVolumeStartWith3(volumeId) {
		return ClearLunPrInfo(volumeId, h.nodeIp, h.nvmePRExecProcessor)
	} else {
		return ClearLunPrInfo(volumeId, h.nodeIp, h.prExecProcessor)
	}
}

func ClearLunPrInfo(volumeId, nodeIp string, processor Processor) error {
	device, err := common.GetDevicePath(volumeId)
	if err != nil {
		if common.PathNotFoundError(err) {
			return nil
		}
		smslog.Infof(err.Error())
		return err
	}
	var prInfo *mpathpersist.PersistentReserve
	if utils.CheckNvmeVolumeStartWith3(volumeId) {
		prInfo, err = nvme.GetPrInfo(device)
	} else {
		prInfo, err = mpathpersist.GetPrInfo(device)
	}
	if err != nil {
		smslog.Errorf("cleanLunPrInfo: failed for get device %s pr info %s", device, err.Error())
		return err
	}

	if len(prInfo.Keys) == 0 {
		smslog.Infof("no need to clear disk [%s] pr ", device)
		return nil
	}

	thisNodePrKey := common.IpV4ToPrKey(nodeIp)
	var alreadyRegistered = false
	for key, _ := range prInfo.Keys {
		if key == thisNodePrKey {
			alreadyRegistered = true
			break
		}
	}
	if !alreadyRegistered {
		param := message.PrRegisterCmdParam{
			RegisterKey: thisNodePrKey,
		}
		_, err = processor.Process(&message.PrCmd{
			CmdType:    message.PrRegister,
			VolumeType: common.MultipathVolume,
			VolumeId:   volumeId,
			CmdParam:   &param,
		})
		if err != nil {
			return fmt.Errorf("ClearLunPrInfo: PrRegisterCmd exec err %s", err.Error())
		}
	}

	param := message.PrClearCmdParam{
		RegisterKey: thisNodePrKey,
	}
	_, err = processor.Process(&message.PrCmd{
		CmdType:    message.PrClear,
		VolumeType: common.MultipathVolume,
		VolumeId:   volumeId,
		CmdParam:   &param,
	})
	if err != nil {
		return fmt.Errorf("ClearLunPrInfo: PrClearCmd exec err %s", err.Error())
	}
	return nil
}
