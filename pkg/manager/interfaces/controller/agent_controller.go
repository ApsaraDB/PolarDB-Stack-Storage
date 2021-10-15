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


package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/application/assembler"
	"polardb-sms/pkg/manager/application/service"
)

type AgentController struct {
	as *service.AgentService
}

func NewAgentController() *AgentController {
	return &AgentController{as: service.NewAgentService()}
}

// @Summary 心跳报告
// @Tags Agent 管理
// @version 1.0
// @Description 用于 心跳报告
// @Accept  json
// @Produce  json
// @Success 200 object string 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /agents/heartbeat [post]
func (c *AgentController) Heartbeat(ctx *gin.Context) {
	//smslog.Debugf("call agent heartbeat")
	heartbeatReq := &assembler.HeartbeatRequest{}
	_ = ParseParam(ctx, heartbeatReq)
	heartbeatResp, err := c.as.Heartbeat(heartbeatReq)
	if err != nil {
		smslog.Errorf("Could not heartbeat: %v", err)
		ReturnError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, heartbeatResp)
}
