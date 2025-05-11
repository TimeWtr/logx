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
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"golang.org/x/sync/errgroup"
	"hash/crc32"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	WalFile = "wal.log"
	// ChunkSize 每次缓存的数据快大小(4KB)，减少碎片写入
	ChunkSize = 1024 * 4
	// CheckSumSize 校验码的长度为32位，4字节
	CheckSumSize = 4
)

// bufferWriterPool buffer专用的缓存池
var bufferWriterPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, ChunkSize))
	},
}

var bytesPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, ChunkSize+ChunkSize)
	},
}

// 生成全局唯一的crc表，简化crc的256种8位计算和查找，一个crc32表占用1KB内存空间
var crc32Table = crc32.MakeTable(crc32.IEEE)

// BufferWriter 使用双缓冲+WAL机制，双缓冲机制最大程度的提高写入日志处理效率
// WAL机制保证日志写入的可靠性，尽可能降低日志数据丢失的可能，ErrorLevel及以上级别
// 的日志只支持同步写入，比如文件立即刷盘，ErrorLevel以下为异步写入
type BufferWriter struct {
	// 当前的缓冲通道
	currentBuffer *bytes.Buffer
	// 执行异步写入操作的缓冲通道
	asyncFlushBuffer *bytes.Buffer
	// 加缓存通道锁保护
	bufferLock *sync.RWMutex
	// 加文件锁保护
	fileLock *sync.Mutex
	// 加全局锁保护写入器管理中心
	stateLock *sync.Mutex
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

func NewBufferWriter(logDir string, interval time.Duration) (*BufferWriter, error) {
	walFile, err := os.OpenFile(filepath.Join(logDir, WalFile), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	ctxl, cancel := context.WithCancel(ctx)
	bw := &BufferWriter{
		currentBuffer:    bufferWriterPool.Get().(*bytes.Buffer),
		asyncFlushBuffer: bufferWriterPool.Get().(*bytes.Buffer),
		bufferLock:       new(sync.RWMutex),
		fileLock:         new(sync.Mutex),
		stateLock:        new(sync.Mutex),
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

// prepareAsyncSwrapData 通道切换前的预准备阶段，需要先拷贝，然后再切换，并返回带crc校验码的字节数组
func (b *BufferWriter) prepareAsyncSwrapData() []byte {
	// 深拷贝防止写入过程中数据污染
	requiredLen := b.currentBuffer.Len() + CheckSumSize
	dataToPersist := bytesPool.Get().([]byte)
	if cap(dataToPersist) < requiredLen {
		dataToPersist = make([]byte, requiredLen)
	}
	dataToPersist = dataToPersist[:requiredLen]
	copy(dataToPersist, b.currentBuffer.Bytes())

	// 缓存通道切换
	b.currentBuffer, b.asyncFlushBuffer = b.asyncFlushBuffer, b.currentBuffer
	b.currentBuffer.Reset()

	// 计算crc校验码
	checkSum := crc32.Checksum(dataToPersist[:len(dataToPersist)-CheckSumSize], crc32Table)
	binary.BigEndian.PutUint32(dataToPersist[len(dataToPersist)-CheckSumSize:], checkSum)
	return dataToPersist
}

// swrapBuffer 用于交换缓冲区和记录WAL写入点，批量写入WAL时也需要生成各自的校验码
// 与同步生成不一样，同步是对每一条写入的日志生成单独的crc校验码，计算开销较大
// 异步生成是针对缓冲区中的4KB批量日志数据进行统一生成唯一校验码，节省计算开销
func (b *BufferWriter) swrapBuffer() error {
	data := b.prepareAsyncSwrapData()
	defer func() {
		data = data[:cap(data)]
		data = data[:0]
		bytesPool.Put(data)
	}()

	var finalErr error
	const MaxRetry = 5
	rand.Seed(time.Now().UnixNano())
	baseDelay := time.Millisecond * 100

	for counter := 0; counter < MaxRetry; counter++ {
		b.fileLock.Lock()
		finalErr = b.writeAndFlush(data)
		b.fileLock.Unlock()
		if finalErr == nil {
			return nil
		}

		// 指数+随机抖动重试策略
		delay := baseDelay * (1 << counter)
		jitter := time.Duration(rand.Int63n(int64(delay / 2)))
		time.Sleep(delay + jitter)
	}

	return fmt.Errorf("failed to swrap buffer and write wal: %v\n", finalErr)
}

func (b *BufferWriter) writeAndFlush(data []byte) error {
	_, err := b.wal.Write(data)
	if err == nil {
		// 强制同步刷盘
		if err = b.sync(); err == nil {
			return nil
		}
		_, _ = os.Stderr.WriteString(fmt.Sprintf("failed to swrap, err: %v\n", err))
	}

	return err
}

// sync 持久化数据到WAL文件
// 1. 通过bufio提供的Flush方法将缓冲区的日志数据刷新到操作系统的PageCache
// 2. 调用底层的文件Sync方法，将PageCache持久化到WAL文件(磁盘)
func (b *BufferWriter) sync() error {
	if err := b.wal.Flush(); err != nil {
		return err
	}

	return b.walFile.Sync()
}

// SyncWrite 同步写入日志数据，同步调用只适用于ErrorLevel及以上级别，确保关键数据不丢失
// 同步日志写入WAL前生成crc校验码，校验码使用大顶端来序列化，兼容网络传输。
func (b *BufferWriter) SyncWrite(data []byte) error {
	newData := b.prepareSyncData(data)
	b.fileLock.Lock()
	defer b.fileLock.Unlock()
	n, err := b.wal.Write(newData)
	if err != nil {
		return err
	}

	if n != len(data)+CheckSumSize {
		return fmt.Errorf("sync write buffer only wrote %d of %d bytes\n", n, len(data))
	}

	return b.sync()
}

// prepareSyncData 同步日志的数据准备阶段，添加CRC校验码
func (b *BufferWriter) prepareSyncData(data []byte) []byte {
	newData := make([]byte, len(data)+CheckSumSize)
	copy(newData[:len(data)], data)

	checksum := crc32.Checksum(data, crc32Table)
	binary.BigEndian.PutUint32(newData[len(data):], checksum)

	return newData
}

// AsyncWrite 异步写入日志数据
func (b *BufferWriter) AsyncWrite(data []byte) error {
	b.bufferLock.Lock()
	defer b.bufferLock.Unlock()
	if b.currentBuffer.Len()+len(data) >= ChunkSize {
		_ = b.swrapBuffer()
	}
	n, err := b.currentBuffer.Write(data)
	if err != nil {
		return err
	}

	if n != len(data) {
		return fmt.Errorf("async write buffer only wrote %d of %d bytes\n", n, len(data))
	}

	return nil
}

// asyncWorker 异步刷新
func (b *BufferWriter) asyncWorker() {
	for range b.ticker.C {
		b.fileLock.Lock()
		if err := b.wal.Flush(); err != nil {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("failed to flush wal file: %v\n", err))
		} else {
			if err = b.walFile.Sync(); err != nil {
				_, _ = os.Stderr.WriteString(fmt.Sprintf("failed to persistent wal file: %v\n", err))
			}
		}
		b.fileLock.Unlock()
	}
}

// AddWriter 动态注册写入器
func (b *BufferWriter) AddWriter(key string, writer Writer) {
	b.stateLock.Lock()
	defer b.stateLock.Unlock()
	b.operator[key] = writer
}

// RemoveWriter 动态删除写入器
func (b *BufferWriter) RemoveWriter(key string) {
	b.stateLock.Lock()
	defer b.stateLock.Unlock()
	delete(b.operator, key)
}

// Close 关闭BufferWriter，释放资源，需要立即执行一次刷盘
func (b *BufferWriter) Close() {
	b.cancel()
	b.ticker.Stop()
	_ = b.sync()
	_ = b.walFile.Close()
	// 重置缓存通道资源
	b.currentBuffer.Reset()
	b.asyncFlushBuffer.Reset()
	bufferWriterPool.Put(b.currentBuffer)
	bufferWriterPool.Put(b.asyncFlushBuffer)
}
