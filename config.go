// Copyright 2025 TimeWtr
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logx

import "github.com/TimeWtr/logx/core"

type Config struct {
	// 日志文件的保存路径
	filePath string
	// 日志级别
	level core.LoggerLevel
	// 文件名称
	filename string
	// 是否打印行号，默认打印
	enableLine bool
	// 是否开启颜色，默认关闭
	enableColor bool
	// ErrorLevel、PanicLevel和FatalLevel级别下，堆栈追踪的行数，即追踪的调用级别，默认3级
	callSkip int
	// 是否开启异步写入
	enableAsync bool
	// 时区
	location string
	// 单个日志文件阈值，允许保存多大的文件，单位bytes
	threshold int64
	// 日志文件的保存周期，单位为天，默认为30天
	period int
	// 历史的日志文件是否开启压缩
	enableCompress bool
	// 压缩的级别
	compressionLevel CompressLevel
}
