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
	"fmt"
	"net/http"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/application/assembler"
	"polardb-sms/pkg/manager/domain/workflow"

	"github.com/gin-gonic/gin"
)

type WorkflowController struct {
	workflowRepo workflow.WorkflowRepository
	as           assembler.WorkflowAssembler
}

// @Summary 查询 WorkflowEntity
// @Tags Workflow 管理
// @version 1.0
// @Description 用于查询 WorkflowEntity
// @Accept  json
// @Produce  json
// @Param workflowId query string true "请求参数"
// @Success 200 object view.WorkflowResponse 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /workflows/:workflowId [get]
func (controller *WorkflowController) FindWorkflowById(ctx *gin.Context) {
	smslog.Infof("call FindWorkflowById")

	workflowId, exist := ctx.Params.Get("workflowId")
	if !exist {
		err := fmt.Errorf("request param not exist workflowId")
		smslog.Errorf(err.Error())
		ReturnError(ctx, err)
		return
	}
	workflow, err := controller.workflowRepo.FindByWorkflowId(workflowId)
	if err != nil {
		smslog.Errorf("Could not query workflow by %s: %v", workflowId, err)
		ReturnError(ctx, err)
		return
	}
	resp := controller.as.ToWorkflowView(workflow)
	ctx.JSON(http.StatusOK, resp)
}

func NewClusterTaskController() *WorkflowController {
	clusterTaskController := &WorkflowController{
		workflowRepo: workflow.NewWorkflowRepository(),
		as:           assembler.NewWorkflowAssembler(),
	}
	return clusterTaskController
}
