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


package view

type WorkflowResponse struct {
	WorkflowId string `json:"workflow_id"`
	Type       int    `json:"type"`
	Step       int    `json:"step"`
	Stages     string `json:"stages"`
	Status     int    `json:"status"`
	Mode       int    `json:"mode"`
}

type WorkflowIdResponse struct {
	WorkflowId string `json:"workflow_id"`
}

