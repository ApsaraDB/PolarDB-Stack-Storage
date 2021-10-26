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
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/application/view"
)

const (
	idempotentKey = "Idempotence-Key"
)

func ParseParam(ctx *gin.Context, param interface{}) error {
	var errResult view.ErrorResult
	if err := common.IoStreamToStruct(ctx.Request.Body, param); err != nil {
		smslog.Errorf("Could not parse request param: %v", err)
		errResult.Err = err.Error()
		ctx.JSON(http.StatusBadRequest, errResult)
		ctx.Abort()
		return err
	}
	return nil
}

func GetTraceContextFromHeader(ctx *gin.Context) common.TraceContext {
	var idempotentId string
	sequences := ctx.Request.Header[idempotentKey]
	if len(sequences) == 0 || len(sequences[0]) == 0 {
		idempotentId = ""
	} else {
		idempotentId = sequences[0]
	}
	return common.NewTraceContext(map[string]string{"traceId": idempotentId})
}
