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

package core

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/TimeWtr/logx/_const"
)

const (
	DefaultParts = 4
	DefaultSkip  = 2
)

type CallWrapOptions func(*CallEntityWrap)

func WithPC() CallWrapOptions {
	return func(w *CallEntityWrap) {
		w.enablePC.Store(true)
	}
}

func WithSkip(skip int32) CallWrapOptions {
	return func(w *CallEntityWrap) {
		w.skip.Store(skip)
	}
}

func WithParts(parts int32) CallWrapOptions {
	return func(w *CallEntityWrap) {
		w.parts.Store(parts)
	}
}

// funcNameCache 全局的方法与PC映射关系缓存，可以显著提高性能
// 正常情况下方法的PC是不会变化的，动态插件例外。
var funcNameCache sync.Map

// callerEntityPool 堆栈实体对象池，减少每次调用堆栈时的对象创建开销和GC开销
var callerEntityPool = sync.Pool{
	New: func() interface{} {
		return &CEntity{}
	},
}

type CallEntityWrap struct {
	// 是否启用函数方法打印
	enablePC atomic.Bool
	// 堆栈信息的级别，打印几级
	skip atomic.Int32
	// 文件路径打印几部分
	parts atomic.Int32
}

func newCallEntityWrap(opts ...CallWrapOptions) *CallEntityWrap {
	cew := &CallEntityWrap{}
	cew.enablePC.Store(false)
	cew.skip.Store(DefaultSkip)
	cew.parts.Store(DefaultParts)

	for _, opt := range opts {
		opt(cew)
	}

	return cew
}

// Fullname 获取条完整的格式化堆栈信息，用于DebugLevel、InfoLevel和WarnLevel
// 单条的堆栈信息不需要指定级别，固定是2.
func (cw *CallEntityWrap) Fullname() string {
	const skip = 2
	ce := newCallerEntity()
	defer ce.release()

	ce.caller(skip)
	if cw.enablePC.Load() {
		return ce.fullstrWithFunc(int(cw.parts.Load()))
	}

	return ce.fullstr(int(cw.parts.Load()))
}

// Fullnames 获取多条完整的格式化堆栈信息，用于ErrorLevel、PanicLevel和FatalLevel
// 多条的堆栈信息必须指定打印的指定级别，需要更多的还原错误异常现场，默认是打印3级别
func (cw *CallEntityWrap) Fullnames() []string {
	ce := newCallerEntity()
	defer ce.release()

	cs, n := ce.callers(int(cw.skip.Load()))
	var res []string
	for i := 0; i < n; i++ {
		pc := cs[i]
		file, line, ok := ce.information(pc)
		if !ok {
			return nil
		}

		ce.ok, ce.pc, ce.file, ce.line = ok, pc, file, line
		if cw.enablePC.Load() {
			res = append(res, ce.fullstrWithFunc(int(cw.parts.Load())))
		} else {
			res = append(res, ce.fullstr(int(cw.parts.Load())))
		}
		ce.release()
	}

	return res
}

// OrignalEntity 获取堆栈的原始数据
func (cw *CallEntityWrap) OrignalEntity() CallerEntity {
	ce := newCallerEntity()
	defer ce.release()

	ce.caller(int(cw.skip.Load()))

	return CallerEntity{
		pc:   ce.pc,
		file: ce.file,
		line: ce.line,
		ok:   ce.ok,
	}
}

// CEntity 堆栈调用实体
type CEntity struct {
	CallerEntity
	// 加锁保护
	lock sync.Mutex
}

type CallerEntity struct {
	// 指向调用的下一级函数
	pc uintptr
	// 调用发生的源文件
	file string
	// 调用发生的源文件行号
	line int
	// 是否成功获取调用的堆栈信息
	ok bool
}

func newCallerEntity() *CEntity {
	obj, _ := callerEntityPool.Get().(*CEntity)
	return obj
}

// release 释放对象
func (c *CEntity) release() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.pc, c.file, c.line, c.ok = 0, "", 0, false
	callerEntityPool.Put(c)
}

// fname 指针指向的方法名称
// 预先从缓存中加载PC与名称，如果查询不到再解析名称，并缓存映射关系
func (c *CEntity) fname() string {
	if !c.ok {
		return _const.Unknown
	}

	fn, ok := funcNameCache.Load(c.pc)
	if ok {
		fname, _ := fn.(string)
		return fname
	}

	fn = runtime.FuncForPC(c.pc).Name()
	fname, _ := fn.(string)
	fnSli := strings.Split(fname, ".")
	if len(fnSli) == 0 {
		return _const.Unknown
	}
	name := fnSli[len(fnSli)-1]
	funcNameCache.Store(c.pc, name)

	return name
}

// caller 捕获堆栈信息
func (c *CEntity) caller(skip int) {
	pc, file, line, ok := runtime.Caller(skip)
	c.lock.Lock()
	defer c.lock.Unlock()

	c.ok, c.pc, c.file, c.line = ok, pc, file, line
}

// fullstr 返回完整的字符串格式数据，不包括方法名
func (c *CEntity) fullstr(parts int) string {
	if !c.ok {
		return _const.Unknown
	}

	var builder strings.Builder
	builder.WriteString(c.getFile(parts))
	builder.WriteString(" line:")
	builder.WriteString(strconv.Itoa(c.line))

	return builder.String()
}

// fullstrWithFunc 返回完整的字符串格式数据，不包括方法名
func (c *CEntity) fullstrWithFunc(parts int) string {
	if !c.ok {
		return "UNKNOWN"
	}

	var builder strings.Builder
	builder.WriteString(c.fname())
	builder.WriteString(c.getFile(parts))
	builder.WriteString(" line:")
	builder.WriteString(strconv.Itoa(c.line))
	builder.WriteString(" func:")
	builder.WriteString(c.fname())

	return builder.String()
}

func (c *CEntity) getFile(parts int) string {
	var file string
	sli := strings.Split(c.file, string(os.PathSeparator))
	if len(sli) == 0 {
		file = c.file
	} else {
		file = filepath.Join(sli[len(sli)-parts:]...)
	}

	return file
}

// callers 捕获多级的堆栈信息
func (c *CEntity) callers(skips int) (pcs []uintptr, cs int) {
	pcs = make([]uintptr, skips)
	c.lock.Lock()
	defer c.lock.Unlock()

	return pcs, runtime.Callers(skips, pcs)
}

// information 根据pc获取详细堆栈信息
func (c *CEntity) information(pc uintptr) (file string, line int, ok bool) {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "UNKNOWN", 0, false
	}

	file, line = fn.FileLine(pc)
	return file, line, true
}
