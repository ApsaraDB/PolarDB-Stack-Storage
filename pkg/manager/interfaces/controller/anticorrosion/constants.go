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

//pvc annotation
const (
	ClusterName                   = "apsara.metric.ppas_name"
	ClusterAction                 = "apsara.metric.ppas_action"
	ClusterActionID               = "apsara.metric.ppas_action_id"
	ClusterRwPod                  = "apsara.metric.ppas_rw_node"
	ClusterRoPods                 = "apsara.metric.ppas_ro_nodes"
	VolumeLockNodeId              = "apsara.metric.lock_node_id"
	K8sAddedCsiPlugin             = "volume.beta.kubernetes.io/storage-provisioner"
	PreProvisionedVolumeWWID      = "apsara.metric.pre_provisioned_volume_wwid"
	PreProvisionedVolumeFormat    = "apsara.metric.pre_provisioned_volume_format"
	PreProvisionedVolumeFormatted = "apsara.metric.pre_provisioned_volume_formatted"
	PreProvisionedVolumeFinalizer = "kubernetes.io/pre-provisioned-volume"
	PvcProcessStatus              = "apsara.metric.pvc_process_status"
)

const (
	StorageNamespace = "kube-system"
	StorageClassName = "csi-polardb-fc"
	StorageLeaderKey = "puresoft-storage-controller-leader"
)

const (
	DefaultFsType = "ext4"
)

const (
	PolarBoxInstanceRW = "apsara.metric.ppas_rw_node"
	PolarBoxInstanceRO = "apsara.metric.ppas_ro_nodes"
)

const (
	SwitchModeRo string = "ro"
	SwitchModeRw string = "rw"

	SwitchStatusSwitching string = "switching"
	SwitchStatusSuccess   string = "success"
	SwitchStatusFailed    string = "failed"
)
