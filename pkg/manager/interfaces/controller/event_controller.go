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
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/application/service"
	serviceAnti "polardb-sms/pkg/manager/application/service/anticorrosion"
	"polardb-sms/pkg/protocol"
)

type EventController struct {
	es *service.EventUploadService
}

func NewEventController() *EventController {
	return &EventController{
		es: service.NewEventUploadService(),
	}
}

// @Summary Event Upload
// @Tags Event 接入管理
// @version 1.0
// @Description 用于存储事件上报
// @Accept  json
// @Produce  json
// @Param name query protocol.Event true "请求参数"
// @Success 200 object string 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /events [post]
func (c *EventController) Upload(ctx *gin.Context) {
	smslog.Infof("call Upload Event")

	var (
		err   error
		event protocol.Event
	)

	if err = common.IoStreamToStruct(ctx.Request.Body, &event); err != nil {
		smslog.Errorf("Could not parse request param: %v", err)
		ReturnError(ctx, err)
		return
	}

	if err = c.es.UploadEvent(&event); err != nil {
		smslog.Errorf("Could not upload event %v: %v", event, err)
		ReturnError(ctx, err)
		return
	}
	ReturnBool(ctx, true)
}

// @Summary Batch Event Upload
// @Tags Event 接入管理
// @version 1.0
// @Description 用于存储事件上报
// @Accept  json
// @Produce  json
// @Param name query protocol.BatchEvent true "请求参数"
// @Success 200 object string 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /events/batch [post]
func (c *EventController) BatchUpload(ctx *gin.Context) {
	smslog.Infof("call BatchUpload Event")

	var (
		err   error
		event protocol.BatchEvent
	)

	if err = common.IoStreamToStruct(ctx.Request.Body, &event); err != nil {
		smslog.Errorf("Could not parse request param: %v", err)
		ReturnError(ctx, err)
		return
	}

	serviceAnti.BatchUpdateByEvents(&event)
	if err = c.es.UploadEvents(event.Events); err != nil {
		smslog.Errorf("Could not upload event %v: %v", event, err)
		ReturnError(ctx, err)
		return
	}
	ReturnBool(ctx, true)
}
