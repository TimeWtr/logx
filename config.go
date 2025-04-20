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

type Config struct {
	// 日志级别
	level LoggerLevel
	// 文件名称
	filename string
	// 是否打印行号
	enableLine bool
	// 是否开启颜色
	enableColor bool
	// ErrorLevel、PanicLevel和FatalLevel级别下，堆栈追踪的行数，即追踪的调用级别
	callSkip int
	// 是否开启异步写入
	enableAsync bool
	// 时区
	location string
}
