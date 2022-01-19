/*
 *    Copyright 2021 chenquan
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package xlog

import (
	"github.com/yunqi/lighthouse/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

var logger = zap.NewNop()

// LoggerWithField release fields to a new logger.
// Plugins can use this method to release plugin name field.
func LoggerWithField(fields ...zap.Field) *zap.Logger {
	return logger.With(fields...)
}

// LoggerModule release fields to a new logger.
// Plugins can use this method to release plugin name field.
func LoggerModule(moduleName string) *zap.Logger {
	return logger.With(zap.String("moduleName", moduleName))
}
func InitLogger(c *config.Log) (err error) {
	var logLevel zapcore.Level
	err = logLevel.UnmarshalText([]byte(c.Level))
	if err != nil {
		return
	}

	hook := &lumberjack.Logger{
		Filename:   c.Filename,   // 日志文件路径
		MaxSize:    c.MaxSize,    // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: c.MaxBackups, // 日志文件最多保存多少个备份
		MaxAge:     c.MaxAge,     // 文件最多保存多少天
		Compress:   c.Compress,   // 是否压缩
	}
	var core zapcore.Core
	if c.Format == "json" {
		core = zapcore.NewCore(zapcore.NewJSONEncoder(
			zap.NewProductionEncoderConfig()),
			zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(hook)),
			logLevel)
	} else if c.Format == "text" {
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
			zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(hook)), logLevel)
	} else {
		core = zapcore.NewNopCore()
	}
	logger = zap.New(core, zap.AddStacktrace(zap.ErrorLevel), zap.AddCaller())
	return
}
