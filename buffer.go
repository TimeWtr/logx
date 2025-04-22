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
	"bufio"
	"bytes"
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"math/rand"
	"os"
	"sync"
	"time"
)

const (
	WalFile = "wal.log"
	// ChunkSize 每次缓存的数据快大小(4KB)，减少碎片写入
	ChunkSize = 1024 * 4
)

var bufferWriterPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, ChunkSize))
	},
}

// BufferWriter 使用双缓冲+WAL机制，双缓冲机制最大程度的提高写入日志处理效率
// WAL机制保证日志写入的可靠性，尽可能降低日志数据丢失的可能，ErrorLevel及以上级别
// 的日志只支持同步写入，比如文件立即刷盘，ErrorLevel以下为异步写入
type BufferWriter struct {
	// 当前的缓冲通道
	currentBuffer *bytes.Buffer
	// 执行异步写入操作的缓冲通道
	asyncFlushBuffer *bytes.Buffer
	// 加锁保护
	bufferLock *sync.RWMutex
	// 多扇出写入器管理中心，用于多种Writer的管理，比如：文件、ES、终端等
	operator map[string]Writer
	// WAL文件缓冲封装
	wal *bufio.Writer
	// WAL文件句柄
	walFile *os.File
	// goroutine管理
	eg errgroup.Group
	// 上下文管理
	ctx context.Context
	// 级联取消
	cancel context.CancelFunc
	// 定时ticker
	ticker *time.Ticker
}

func NewBufferWriter(interval time.Duration) (*BufferWriter, error) {
	walFile, err := os.Create(WalFile)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	ctxl, cancel := context.WithCancel(ctx)
	bw := &BufferWriter{
		currentBuffer:    bufferWriterPool.Get().(*bytes.Buffer),
		asyncFlushBuffer: bufferWriterPool.Get().(*bytes.Buffer),
		bufferLock:       new(sync.RWMutex),
		operator:         make(map[string]Writer),
		wal:              bufio.NewWriterSize(walFile, ChunkSize),
		walFile:          walFile,
		eg:               errgroup.Group{},
		ctx:              ctxl,
		cancel:           cancel,
		ticker:           time.NewTicker(interval),
	}

	// 开启定时任务异步执行刷盘
	go bw.asyncWorker()

	return bw, nil
}

// SwrapBuffer 用于交换缓冲区和记录WAL写入点
func (b *BufferWriter) SwrapBuffer() error {
	b.bufferLock.Lock()
	// 深拷贝防止写入过程中数据污染
	dataToPersist := make([]byte, b.asyncFlushBuffer.Len())
	copy(dataToPersist, b.asyncFlushBuffer.Bytes())

	b.currentBuffer, b.asyncFlushBuffer = b.asyncFlushBuffer, b.currentBuffer
	b.currentBuffer.Reset()
	b.bufferLock.Unlock()

	var finalErr error
	const MaxRetry = 5
	rand.Seed(time.Now().UnixNano())

	baseDelay := time.Millisecond * 100
	for counter := 0; counter < MaxRetry; counter++ {
		_, err := b.wal.Write(dataToPersist)
		if err == nil {
			// 强制同步刷盘
			if err = b.sync(); err == nil {
				return nil
			}
			_, _ = os.Stderr.WriteString(fmt.Sprintf("failed to swrap, err: %v", err))
			finalErr = err
		}

		// 指数+随机抖动重试策略
		delay := baseDelay * (1 << counter)
		jitter := time.Duration(rand.Int63n(int64(delay / 2)))
		time.Sleep(delay + jitter)
	}

	return fmt.Errorf("failed to swrap buffer and write wal: %v", finalErr)
}

// sync 持久化数据到WAL文件
// 1. 通过bufio提供的Flush方法将缓冲区的日志数据刷新到操作系统的PageCache
// 2. 调用底层的文件Sync方法，将PageCache持久化到WAL文件(磁盘)
func (b *BufferWriter) sync() error {
	if err := b.wal.Flush(); err == nil {
		return nil
	}

	return b.walFile.Sync()
}

// SyncWrite 同步写入日志数据，同步调用只适用于ErrorLevel及以上级别，确保关键数据不丢失
func (b *BufferWriter) SyncWrite(data []byte) error {
	n, err := b.wal.Write(data)
	if err != nil {
		return err
	}

	if n != len(data) {
		return fmt.Errorf("sync write buffer only wrote %d of %d bytes", n, len(data))
	}

	return b.sync()
}

// AsyncWrite 异步写入日志数据
func (b *BufferWriter) AsyncWrite(data []byte) error {
	if b.currentBuffer.Len()+len(data) >= ChunkSize {
		go func() {
			_ = b.SwrapBuffer()
		}()
	}

	n, err := b.currentBuffer.Write(data)
	if err != nil {
		return err
	}

	if n != len(data) {
		return fmt.Errorf("async write buffer only wrote %d of %d bytes", n, len(data))
	}

	return nil
}

// AddWriter 动态注册写入器
func (b *BufferWriter) AddWriter(key string, writer Writer) {
	b.bufferLock.Lock()
	defer b.bufferLock.Unlock()
	b.operator[key] = writer
}

// RemoveWriter 动态删除写入器
func (b *BufferWriter) RemoveWriter(key string) {
	b.bufferLock.Lock()
	defer b.bufferLock.Unlock()
	delete(b.operator, key)
}

// asyncWorker 异步刷新
func (b *BufferWriter) asyncWorker() {
	for range b.ticker.C {
		if err := b.wal.Flush(); err == nil {
			if err = b.walFile.Sync(); err != nil {
				_, _ = os.Stderr.WriteString(fmt.Sprintf("failed to persistent wal file: %v", err))
			}
		} else {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("failed to flush wal file: %v", err))
		}
	}
}

// Close 关闭BufferWriter，释放资源，需要立即执行一次刷盘
func (b *BufferWriter) Close() {
	b.cancel()
	_ = b.sync()
	_ = b.walFile.Close()
	b.ticker.Stop()
}
