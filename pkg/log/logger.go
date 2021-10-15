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
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"runtime/debug"
	"strings"
)

const (
	MaxSize    = 50 //50MiB
	MaxBackups = 20
	MaxAge     = 30 //30 day
)

// Level type
type Level uint32

type Logger interface {
	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})

	Info(args ...interface{})
	Debug(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
}

var _logger *zap.SugaredLogger
var _smsZapLogger *smsZapLogger

func Infof(format string, args ...interface{}) {
	_logger.Infof(format, args...)
}

func Debugf(format string, args ...interface{}) {
	_logger.Debugf(format, args...)
}

func Warnf(format string, args ...interface{}) {
	_logger.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	_logger.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	_logger.Fatalf(format, args...)
}

func Info(args ...interface{}) {
	_logger.Info(args...)
}

func Debug(args ...interface{}) {
	_logger.Debug(args...)
}

func Warn(args ...interface{}) {
	_logger.Warn(args...)
}

func Error(args ...interface{}) {
	_logger.Error(args...)
}

func Fatal(args ...interface{}) {
	_logger.Fatal(args...)
}

func WithContext(ctx map[string]string) Logger {
	if ctx != nil {
		_smsZapLogger.ctx = ctx
	}
	return _smsZapLogger
}

func LogLevel(level string) zapcore.Level {
	level = strings.ToUpper(level)
	switch level {
	case "DEBUG":
		return zapcore.DebugLevel
	case "INFO":
		return zapcore.InfoLevel
	case "WARN":
		return zapcore.WarnLevel
	case "ERROR":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func Flush() {
	_ = _logger.Sync()
}

func LogPanic() {
	if err := recover(); err != nil {
		Infof("panic info %v, stack %s", err, string(debug.Stack()))
		Flush()
	}
}