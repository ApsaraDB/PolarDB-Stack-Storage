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
	"github.com/gin-gonic/gin"
	"net/http"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/application/service"
	"polardb-sms/pkg/manager/application/view"
)

type PrCheckController struct {
	cs *service.PrCheckService
}

func NewPrCheckController() *PrCheckController {
	return &PrCheckController{cs: service.NewPrCheckService()}
}

// @Summary Pr Check Detail Capability
// @Tags PR Check 接入管理
// @version 1.0
// @Description 用于检查PR所支持的能力
// @Accept  json
// @Produce  json
// @Param checkReq body view.PrCheckRequest true "请求参数"
// @Success 200 object view.WorkflowIdResponse 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /pr-check/detail [post]
func (c *PrCheckController) CheckDetailCapabilities(ctx *gin.Context) {
	smslog.Infof("call CheckDetailCapabilities")
	var prCheckRequest view.PrCheckRequest
	if err := ParseParam(ctx, &prCheckRequest); err != nil {
		smslog.Errorf("Cloud not parse cluster lun pr-check request %v: %v", prCheckRequest, err)
		ReturnError(ctx, err)
		return
	}
	wflResp, err := c.cs.CheckDetail(&prCheckRequest)
	if err != nil {
		smslog.Errorf("Could not check detail capabilities for %v: %v", prCheckRequest, err)
		ReturnError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, wflResp)
}

// @Summary Pr Check Detail Capability Response
// @Tags PR Check 接入管理
// @version 1.0
// @Description 用于检查PR所支持的能力的结果查询
// @Accept  json
// @Produce  json
// @Param volumeId query string true "请求参数"
// @Param volumeClass query string true "请求参数"
// @Success 200 object view.PrCheckResponse 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /pr-check/detail [get]
func (c *PrCheckController) QueryDetailCapabilities(ctx *gin.Context) {
	volumeId, exist := ctx.GetQuery("volumeId")
	if !exist {
		err := fmt.Errorf("request param not exist volumeId")
		smslog.Errorf(err.Error())
		ReturnError(ctx, err)
		return
	}
	volumeClass, exist := ctx.GetQuery("volumeClass")
	if !exist {
		err := fmt.Errorf("request param not exist volumeClass")
		smslog.Errorf(err.Error())
		ReturnError(ctx, err)
		return
	}
	ret, err := c.cs.CheckDetailResponse(volumeId, volumeClass)
	if err != nil {
		smslog.Errorf(err.Error())
		ReturnError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, ret)
}

// @Summary Pr Check Overall Capability
// @Tags PR Check 接入管理
// @version 1.0
// @Description 用于获取PR所支持的能力
// @Accept  json
// @Produce  json
// @Param pvc body view.PrCheckRequest true "请求参数"
// @Success 200 object view.PrCheckResponse 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /pr-check/overall [post]
func (c *PrCheckController) CheckOverallCapabilities(ctx *gin.Context) {
	smslog.Infof("call CheckOverallCapabilities")
	var prCheckRequest view.PrCheckRequest
	if err := ParseParam(ctx, &prCheckRequest); err != nil {
		smslog.Errorf("Cloud not parse cluster lun pr-check request %v: %v", prCheckRequest, err)
		ReturnError(ctx, err)
		return
	}
	ret, err := c.cs.CheckOverall(&prCheckRequest)
	if err != nil {
		smslog.Errorf("Could not check overall capabilities for %v: %v", prCheckRequest, err)
		ReturnError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, ret)
}
