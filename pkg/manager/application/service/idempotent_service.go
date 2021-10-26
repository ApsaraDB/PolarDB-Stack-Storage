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

package service

import (
	"github.com/gin-gonic/gin"
	"net/http"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/domain/repository"
	"strings"
)

const (
	idempotentKey    = "Idempotence-Key"
	UrlPathSplitSign = "/"
	SourceJoinSign   = "_"
)

var (
	filterUrls = []string{"/agent/heartbeat"}
)

func IdempotentCheck(ctx *gin.Context) {
	if ctx.Request.Method == http.MethodPost || ctx.Request.Method == http.MethodDelete {
		for _, url := range filterUrls {
			if ctx.Request.URL.String() == url {
				smslog.Debugf("The request url %v not need to sequence processing, ignore", ctx.Request.URL.String())
				return
			}
		}

		sequences := ctx.Request.Header[idempotentKey]
		if len(sequences) == 0 || len(sequences[0]) == 0 {
			smslog.Debugf("Could not find sequence id in request header: %v", ctx.Request.Header)
			return
		}
		idempotentId := sequences[0]
		uris := strings.Split(ctx.Request.URL.String(), UrlPathSplitSign)
		source := strings.Join(append(uris, ctx.Request.Method), SourceJoinSign)

		existIdempotentKey, err := repository.NewIdempotentRepo().FindBySourceAndId(idempotentId, source)
		if err != nil {
			smslog.Errorf("Could not find existIdempotentKey: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			ctx.Abort()
			return
		}
		smslog.Infof("The existIdempotentKey source is %s: %v", source, existIdempotentKey)
		if existIdempotentKey.Source == source && existIdempotentKey.IdempotentId == idempotentId {
			smslog.Infof("The IdempotentKey exist, return...")
			ctx.JSON(http.StatusOK, gin.H{"workflowId": existIdempotentKey.WorkflowId})
			ctx.Abort()
			return
		}
		ctx.Next()
		maybeWflIds, ok := ctx.Writer.Header()[idempotentKey]
		if !ok {
			ctx.Abort()
			return
		}
		if existIdempotentKey == nil || existIdempotentKey.IdempotentId != idempotentId {
			smslog.Infof("The existIdempotentKey not exist, creating...")
			idt := &repository.Idempotent{
				Source:       source,
				IdempotentId: idempotentId,
				WorkflowId:   maybeWflIds[0],
			}
			if _, err = repository.NewIdempotentRepo().Create(idt); err != nil {
				smslog.Errorf("Could not create existIdempotentKey: %v", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				ctx.Abort()
				return
			}
		}
	}
}
