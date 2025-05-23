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

import (
	"fmt"
	"strings"
	"sync"

	"github.com/TimeWtr/logx/core"
)

type Logger interface {
	Debug(v ...any)
	Info(v ...any)
	Warn(v ...any)
	Error(v ...any)
	Panic(v ...any)
	Fatal(v ...any)
	Debugf(format string, v ...any)
	Infof(format string, v ...any)
	Warnf(format string, v ...any)
	Errorf(format string, v ...any)
	Panicf(format string, v ...any)
	Fatalf(format string, v ...any)
}

const (
	DefaultErrCoreSkip = 3
	DefaultLogSize     = 100 * 1024 * 1024
	DefaultPeriod      = 30
	DefaultLocation    = "Asia/Shanghai"
	DefaultFilename    = "server.log"
)

type WriteMode int

const (
	NormalMode WriteMode = iota
	FormatMode
)

type Log struct {
	// 配置信息
	cfg *Config
	// 并发保护
	mu *sync.Mutex
	// 日志加颜色输出
	cp core.ColorPlugin
}

func NewLog(filePath string, opts ...Options) (Logger, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path can't be empty")
	}

	cfg := &Config{
		filePath:       filePath,
		filename:       DefaultFilename,
		level:          core.InfoLevel,
		location:       DefaultLocation,
		enableLine:     true,
		callSkip:       DefaultErrCoreSkip,
		threshold:      DefaultLogSize,
		period:         DefaultPeriod,
		enableCompress: false,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	l := &Log{
		cfg: cfg,
		mu:  new(sync.Mutex),
		cp:  core.NewANSIColorPlugin(),
	}

	return l, nil
}

func (l *Log) prefix(enabled bool, level core.LoggerLevel, v ...any) string {
	var builder strings.Builder
	builder.WriteString(l.cp.Format(enabled, level))
	builder.WriteString(fmt.Sprint(v...))
	return builder.String()
}

func (l *Log) prefixf(enabled bool, level core.LoggerLevel, format string, v ...any) string {
	var builder strings.Builder
	builder.WriteString(l.cp.Format(enabled, level))
	if level.Prohibit(core.InfoLevel) {
		// TODO 处理这个分支
	}
	builder.WriteString(fmt.Sprintf(format, v...))
	return builder.String()
}

func (l *Log) Debug(v ...any) {
	if l.cfg.level.Prohibit(core.DebugLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.normalExecf(NormalMode, core.DebugLevel, "", v...)
}

func (l *Log) Info(v ...any) {
	if l.cfg.level.Prohibit(core.InfoLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.normalExecf(NormalMode, core.InfoLevel, "", v...)
}

func (l *Log) Warn(v ...any) {
	if l.cfg.level.Prohibit(core.WarnLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.normalExecf(NormalMode, core.WarnLevel, "", v...)
}

func (l *Log) Error(v ...any) {
	if l.cfg.level.Prohibit(core.ErrorLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.abnormalExecf(NormalMode, core.ErrorLevel, "", v...)
}

func (l *Log) Panic(v ...any) {
	if l.cfg.level.Prohibit(core.PanicLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.abnormalExecf(NormalMode, core.PanicLevel, "", v...)
}

func (l *Log) Fatal(v ...any) {
	if l.cfg.level.Prohibit(core.FatalLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.abnormalExecf(NormalMode, core.FatalLevel, "", v...)
}

func (l *Log) Debugf(format string, v ...any) {
	if l.cfg.level.Prohibit(core.DebugLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.normalExecf(FormatMode, core.DebugLevel, format, v...)
}

func (l *Log) Infof(format string, v ...any) {
	if l.cfg.level.Prohibit(core.InfoLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.normalExecf(FormatMode, core.InfoLevel, format, v...)
}

func (l *Log) Warnf(format string, v ...any) {
	if l.cfg.level.Prohibit(core.WarnLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.normalExecf(FormatMode, core.WarnLevel, format, v...)
}

func (l *Log) Errorf(format string, v ...any) {
	if l.cfg.level.Prohibit(core.ErrorLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.abnormalExecf(FormatMode, core.ErrorLevel, format, v...)
}

func (l *Log) Panicf(format string, v ...any) {
	if l.cfg.level.Prohibit(core.PanicLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.abnormalExecf(FormatMode, core.PanicLevel, format, v...)
}

func (l *Log) Fatalf(format string, v ...any) {
	if l.cfg.level.Prohibit(core.FatalLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.abnormalExecf(FormatMode, core.FatalLevel, format, v...)
}

// normalExecf 正常级别下真正执行写入的方法
func (l *Log) normalExecf(mode WriteMode, level core.LoggerLevel, format string, v ...any) {
	var msg string
	switch mode {
	case NormalMode:
		msg = l.prefix(false, level, v...)
	case FormatMode:
		msg = l.prefixf(false, level, format, v...)
	}

	fmt.Println(msg)
}

// abnormalExecf 异常级别下真正执行写入的方法
func (l *Log) abnormalExecf(mode WriteMode, level core.LoggerLevel, format string, v ...any) {
	var msg string
	switch mode {
	case NormalMode:
		msg = l.prefix(true, level, v...)
	case FormatMode:
		msg = l.prefixf(false, level, format, v...)
	}
	fmt.Println(msg)
}

// abnormalStack 用于打印异常情况下的多行堆栈信息，特殊处理，Debug、Info级别不需要
// 返回写入的数据大小
//
//nolint:unused  // 后边要使用
func (l *Log) abnormalStack() int {
	var builder strings.Builder
	//for _, s := range MultiLevel(l.cfg.callSkip) {
	//	str := "\t" + s + "\n"
	//	builder.WriteString(str)
	//}

	res := builder.String()
	//_, _ = l.rs.logout.WriteString(res)
	fmt.Println(res)
	return len(res)
}
