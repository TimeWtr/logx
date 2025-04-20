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
	DeaultLogSize      = 100 * 1024 * 1024
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
	// 轮转策略
	rs *RotateStrategy
	// 日志加颜色输出
	cp ColorPlugin
}

func NewLog(filePath string, opts ...Options) (Logger, error) {
	cfg := &Config{
		filePath:         filePath,
		filename:         DefaultFilename,
		level:            InfoLevel,
		location:         DefaultLocation,
		enableLine:       true,
		callSkip:         DefaultErrCoreSkip,
		threshold:        DeaultLogSize,
		period:           DefaultPeriod,
		enableCompress:   false,
		compressionLevel: DefaultCompression,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.enableCompress && cfg.compressionLevel.valid() {
		return nil, fmt.Errorf("invalid compression level: %d", cfg.compressionLevel)
	}

	rs, err := NewRotateStrategy(cfg.filename, cfg.threshold, cfg.enableCompress, cfg.compressionLevel)
	if err != nil {
		return nil, err
	}
	// 异步执行定时轮转
	go rs.AsyncWork()

	l := &Log{
		cfg: cfg,
		mu:  new(sync.Mutex),
		rs:  rs,
		cp:  NewANSIColorPlugin(cfg.enableColor),
	}

	return l, nil
}

func (l *Log) prefix(level LoggerLevel, v ...any) string {
	var builder strings.Builder
	builder.WriteString(l.cp.Format(level))
	builder.WriteString(Streamline())
	builder.WriteString(fmt.Sprint(v...))
	return builder.String()
}

func (l *Log) prefixf(level LoggerLevel, format string, v ...any) string {
	var builder strings.Builder
	builder.WriteString(l.cp.Format(level))
	if level.check(InfoLevel) {
		builder.WriteString(Streamline())
	}
	builder.WriteString(fmt.Sprintf(format, v...))
	return builder.String()
}

func (l *Log) Debug(v ...any) {
	if l.cfg.level.check(DebugLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.normalExecf(NormalMode, DebugLevel, "", v...)
}

func (l *Log) Info(v ...any) {
	if l.cfg.level.check(InfoLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.normalExecf(NormalMode, InfoLevel, "", v...)
}

func (l *Log) Warn(v ...any) {
	if l.cfg.level.check(WarnLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.normalExecf(NormalMode, WarnLevel, "", v...)
}

func (l *Log) Error(v ...any) {
	if l.cfg.level.check(ErrorLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.abnormalExecf(NormalMode, ErrorLevel, "", v...)
}

func (l *Log) Panic(v ...any) {
	if l.cfg.level.check(PanicLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.abnormalExecf(NormalMode, PanicLevel, "", v...)
}

func (l *Log) Fatal(v ...any) {
	if l.cfg.level.check(FatalLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.abnormalExecf(NormalMode, FatalLevel, "", v...)
}

func (l *Log) Debugf(format string, v ...any) {
	if l.cfg.level.check(DebugLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.normalExecf(FormatMode, DebugLevel, format, v...)
}

func (l *Log) Infof(format string, v ...any) {
	if l.cfg.level.check(InfoLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.normalExecf(FormatMode, InfoLevel, format, v...)
}

func (l *Log) Warnf(format string, v ...any) {
	if l.cfg.level.check(WarnLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.normalExecf(FormatMode, WarnLevel, format, v...)
}

func (l *Log) Errorf(format string, v ...any) {
	if l.cfg.level.check(ErrorLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.abnormalExecf(FormatMode, ErrorLevel, format, v...)
}

func (l *Log) Panicf(format string, v ...any) {
	if l.cfg.level.check(PanicLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.abnormalExecf(FormatMode, PanicLevel, format, v...)
}

func (l *Log) Fatalf(format string, v ...any) {
	if l.cfg.level.check(FatalLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.abnormalExecf(FormatMode, FatalLevel, format, v...)
}

// normalExecf 正常级别下真正执行写入的方法
func (l *Log) normalExecf(mode WriteMode, level LoggerLevel, format string, v ...any) {
	err := l.rs.Rotate()
	if err != nil {
		return
	}

	var msg string
	switch mode {
	case NormalMode:
		msg = l.prefix(level, v...)
	case FormatMode:
		msg = l.prefixf(level, format, v...)
	}

	l.rs.lg.Println(msg)
	l.rs.SetCurrentSize(int64(len(msg)))
}

// abnormalExecf 异常级别下真正执行写入的方法
func (l *Log) abnormalExecf(mode WriteMode, level LoggerLevel, format string, v ...any) {
	err := l.rs.Rotate()
	if err != nil {
		return
	}

	var msg string
	switch mode {
	case NormalMode:
		msg = l.prefix(level, v...)
	case FormatMode:
		msg = l.prefixf(level, format, v...)
	}

	l.rs.lg.Println(msg)
	size := l.abnormalStack() + len(msg)
	l.rs.SetCurrentSize(int64(size))
}

// abnormalStack 用于打印异常情况下的多行堆栈信息，特殊处理，Debug、Info级别不需要
// 返回写入的数据大小
func (l *Log) abnormalStack() int {
	var builder strings.Builder
	for _, s := range MultiLevel(l.cfg.callSkip) {
		str := "\t" + s + "\n"
		builder.WriteString(str)
	}

	res := builder.String()
	_, _ = l.rs.logout.WriteString(res)
	return len(res)
}
