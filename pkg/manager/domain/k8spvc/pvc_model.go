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

type Pvc struct {
	Id           int    `xorm:"not null pk autoincr INT"`
	PvcName      string `xorm:"unique(pvc_name_UNIQUE) VARCHAR(45)"`
	PvcNamespace string `xorm:"unique(pvc_name_UNIQUE) VARCHAR(45)"`
	PvcStatus    string `xorm:"VARCHAR(255)"`
	VolumeClass  string `xorm:"VARCHAR(45)"`
}
