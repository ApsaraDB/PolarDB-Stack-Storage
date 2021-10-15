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
	"polardb-sms/pkg/manager/application/service"
	"polardb-sms/pkg/manager/application/view"
)

type DeviceMapperController struct {
	dms *service.DeviceMapperService
}

// @Summary 生成预览的lv的配置文件
// @Tags LV 管理
// @version 1.0
// @Description 用于生成预览的lv的配置文件
// @Accept  json
// @Produce  json
// @Param clusterLv body view.GenerateDmCreateCmdRequest true "请求参数"
// @Success 200 object view.DmCreateCmdResponse 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /cluster-lvs/dm-conf [post]
func (c *DeviceMapperController) GeneratePreviewConf(ctx *gin.Context) {
	smslog.Infof("call GeneratePreviewConf")
	var genDmCmdRequest view.GenerateDmCreateCmdRequest
	if err := ParseParam(ctx, &genDmCmdRequest); err != nil {
		smslog.Errorf("Cloud not parse cluster lun preview request %v: %v", genDmCmdRequest, err)
		ReturnError(ctx, err)
		return
	}
	previewConf, err := c.dms.GenerateConf(&genDmCmdRequest)
	if err != nil {
		smslog.Errorf("Could not create cluster lv  %v: %v", genDmCmdRequest, err)
		ReturnError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, previewConf)
}

func NewDeviceMapperController() *DeviceMapperController {
	return &DeviceMapperController{
		dms: service.NewDeviceMapperService(),
	}
}
