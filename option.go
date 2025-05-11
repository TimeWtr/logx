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

type Options func(*Config)

// WithColor 是否开启日志输出颜色
func WithColor() Options {
	return func(l *Config) {
		l.enableColor = true
	}
}

// WithLevel 设置日志级别，如果不设置，默认级别是InfoLevel
func WithLevel(level core.LoggerLevel) Options {
	return func(l *Config) {
		l.level = level
	}
}

// WithFileName 配置日志文件名称
func WithFileName(fileName string) Options {
	return func(l *Config) {
		l.filename = fileName
	}
}

// WithLine 开启日志打印行号
func WithLine(enable bool) Options {
	return func(l *Config) {
		l.enableLine = enable
	}
}

// WithCallSkip 设置ErrorLevel、PanicLevel和FatalLevel日志级别时打印的堆栈信息层级
func WithCallSkip(skip int) Options {
	return func(l *Config) {
		l.callSkip = skip
	}
}

// WithAsync 开启异步写入日志文件
func WithAsync() Options {
	return func(l *Config) {
		l.enableAsync = true
	}
}

// WithLocation 设置时区，默认是Asia/Shanghai
func WithLocation(location string) Options {
	return func(l *Config) {
		l.location = location
	}
}

// WithThreshold 设置单个文件的大小，单位为MB，默认为100MB
func WithThreshold(threshold int64) Options {
	return func(l *Config) {
		l.threshold = threshold
	}
}

// WithPeriod 设置日志文件的保存周期，默认周期30天
func WithPeriod(period int) Options {
	return func(l *Config) {
		l.period = period
	}
}

// WithEnableCompress 开启历史日志文件压缩
func WithEnableCompress() Options {
	return func(l *Config) {
		l.enableCompress = true
	}
}

// WithCompressionLevel 设置压缩的级别，如果不设置则为DefaultCompression
func WithCompressionLevel(level CompressLevel) Options {
	return func(l *Config) {
		l.compressionLevel = level
	}
}
