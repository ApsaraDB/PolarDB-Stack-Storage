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


package log

import (
	"fmt"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path"
	"strings"
)

func InitLogger(logDir string, logFile string, logLevel zapcore.Level) {
	filePath := path.Join(logDir, logFile)
	writer := getLogWriter(filePath, logLevel)
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, writer, logLevel)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	_logger = logger.Sugar()
	_smsZapLogger = &smsZapLogger{
		logger: _logger,
		ctx:    nil,
	}
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter(logPath string, logLevel zapcore.Level) zapcore.WriteSyncer {
	hook := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    MaxSize,
		MaxBackups: MaxBackups,
		MaxAge:     MaxAge,
	}
	if logLevel == zapcore.DebugLevel {
		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(hook), zapcore.AddSync(os.Stdout))
	}
	return zapcore.NewMultiWriteSyncer(zapcore.AddSync(hook), zapcore.AddSync(os.Stdout))
	//return zapcore.AddSync(hook)
}

type smsZapLogger struct {
	logger *zap.SugaredLogger
	ctx    map[string]string
}

const ctxTemplate = "[%s] "

func (l *smsZapLogger) ctxString() string {
	if l.ctx != nil {
		var sb strings.Builder
		for k, v := range l.ctx {
			sb.WriteString(fmt.Sprintf(" %s:%s ", k, v))
		}
		return sb.String()
	}
	return ""
}

func (l *smsZapLogger) Infof(format string, args ...interface{}) {
	if l.ctx != nil {
		l.logger.Infof(fmt.Sprintf(ctxTemplate, l.ctxString())+format, args...)
		return
	}
	l.logger.Infof(format, args...)
}

func (l *smsZapLogger) Debugf(format string, args ...interface{}) {
	if l.ctx != nil {
		l.logger.Debugf(fmt.Sprintf(ctxTemplate, l.ctxString())+format, args...)
		return
	}
	l.logger.Debugf(format, args...)
}

func (l *smsZapLogger) Warnf(format string, args ...interface{}) {
	if l.ctx != nil {
		l.logger.Warnf(fmt.Sprintf(ctxTemplate, l.ctxString())+format, args...)
		return
	}
	l.logger.Warnf(format, args...)
}

func (l *smsZapLogger) Errorf(format string, args ...interface{}) {
	if l.ctx != nil {
		l.logger.Errorf(fmt.Sprintf(ctxTemplate, l.ctxString())+format, args...)
		return
	}
	l.logger.Errorf(format, args...)
}

func (l *smsZapLogger) Fatalf(format string, args ...interface{}) {
	if l.ctx != nil {
		l.logger.Fatalf(fmt.Sprintf(ctxTemplate, l.ctxString())+format, args...)
		return
	}
	l.logger.Fatalf(format, args...)
}

func (l *smsZapLogger) Info(args ...interface{}) {
	if l.ctx != nil {
		l.logger.Info(append([]interface{}{fmt.Sprintf(ctxTemplate, l.ctxString())}, args...)...)
		return
	}
	l.logger.Info(args)
}

func (l *smsZapLogger) Debug(args ...interface{}) {
	if l.ctx != nil {
		l.logger.Debug(append([]interface{}{fmt.Sprintf(ctxTemplate, l.ctxString())}, args...)...)
		return
	}
	l.logger.Debug(args)
}

func (l *smsZapLogger) Warn(args ...interface{}) {
	if l.ctx != nil {
		l.logger.Warn(append([]interface{}{fmt.Sprintf(ctxTemplate, l.ctxString())}, args...)...)
		return
	}
	l.logger.Warn(args)
}

func (l *smsZapLogger) Error(args ...interface{}) {
	if l.ctx != nil {
		l.logger.Error(append([]interface{}{fmt.Sprintf(ctxTemplate, l.ctxString())}, args...)...)
		return
	}
	l.logger.Error(args)
}

func (l *smsZapLogger) Fatal(args ...interface{}) {
	if l.ctx != nil {
		l.logger.Fatal(append([]interface{}{fmt.Sprintf(ctxTemplate, l.ctxString())}, args...)...)
		return
	}
	l.logger.Fatal(args)
}

func (l *smsZapLogger) Flush() error {
	return l.logger.Sync()
}
