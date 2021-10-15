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
	"polardb-sms/pkg/manager/application/view"
)

func ReturnError(ctx *gin.Context, err error) {
	var errResult view.ErrorResult
	errResult.Err = err.Error()
	ctx.JSON(http.StatusInternalServerError, errResult)
	ctx.Abort()
}

func ReturnBool(ctx *gin.Context, value bool) {
	var result view.BoolResult
	result.Result = value
	ctx.JSON(http.StatusOK, result)
}