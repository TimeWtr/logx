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
	"errors"
	"sync"
	"time"
)

type BufferStatus uint8

const (
	// WritingStatus 实时写入中
	WritingStatus BufferStatus = iota
	// RejectStatus 拒绝写入
	RejectStatus
)

const (
	// SizeThreshold 缓冲区的切换大小阈值
	SizeThreshold = 1024
	// TimeThreshold 缓冲区切换的时间阈值
	TimeThreshold = time.Second * 5
)

type Buffer struct {
	// 环形缓冲区，包含两个缓冲通道，初始状态下，下标为0的缓冲区为活跃缓冲区，实时接收日志数据
	// 下标为1的缓冲区为备用缓冲区，当下标为0的缓冲区达到阈值/定时，进行缓冲通道的切换，下标1缓冲区
	// 切换为活跃缓冲区，开始实时接收日志数据，下标为0的缓冲区切换为异步刷盘缓冲区，异步从缓冲区中读取
	// 日志数据给到Writer写入器写入日志文件。循环往复，不断切换缓冲区。
	bfs [2]chan string
	// 异步读取通道
	readq chan string
	// 关闭缓冲区的信号
	sig chan struct{}
	// 切换信号
	switchSig chan struct{}
	// 单例
	once sync.Once
	// 活跃通道下标
	index uint8
	// 当前缓冲区状态
	status BufferStatus
	// 活跃缓冲区写入的字节大小
	size uint64
	// 加锁保护
	lock sync.RWMutex
}

func NewBuffer(capacity int64) *Buffer {
	bfs := [2]chan string{}
	bfs[0] = make(chan string, capacity)
	bfs[1] = make(chan string, capacity)

	b := &Buffer{
		bfs:       bfs,
		sig:       make(chan struct{}),
		switchSig: make(chan struct{}, 2),
		readq:     make(chan string, capacity*2),
		status:    WritingStatus,
		lock:      sync.RWMutex{},
	}

	go b.asyncReader()
	go b.asyncWork()

	return b
}

func (b *Buffer) Write(p string) error {
	select {
	case <-b.sig:
		return nil
	default:
	}

	b.lock.RLock()
	if b.status == RejectStatus {
		b.lock.RUnlock()
		return errors.New("buffer is closed")
	}
	b.lock.RUnlock()

	b.lock.Lock()
	defer b.lock.Unlock()
	if b.status == RejectStatus {
		return errors.New("buffer is closed")
	}

	pSize := len(p)
	if b.size+uint64(pSize) > SizeThreshold {
		// 执行切换逻辑
		b.sw()
	}

	select {
	case b.bfs[b.index] <- p:
		b.size += uint64(pSize)
		return nil
	case <-b.sig:
		return errors.New("buffer is closed")
	default:
		return errors.New("write error")
	}
}

func (b *Buffer) Register() <-chan string {
	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.readq
}

// sw 执行切换逻辑
func (b *Buffer) sw() {
	b.index = 1 - b.index
	b.size = 0
	b.switchSig <- struct{}{}
}

func (b *Buffer) asyncWork() {
	ticker := time.NewTicker(TimeThreshold)
	defer ticker.Stop()

	for {
		select {
		case <-b.sig:
			return
		case <-ticker.C:
			//b.lock.Lock()
			//if b.status == RejectStatus {
			//	b.lock.Unlock()
			//	return
			//}
			//
			//// 执行切换逻辑
			//b.sw()
			//b.lock.Unlock()
		}
	}
}

// asyncReader 异步读取器，后台异步的把缓冲通道中的日志数据读取出来，并写入大readq中
func (b *Buffer) asyncReader() {
	for {
		select {
		case <-b.sig:
			close(b.readq)
			return
		case <-b.switchSig:
			b.lock.Lock()
			asyncQ := b.bfs[1-b.index]
			b.lock.Unlock()

			go func() {
				for {
					select {
					case data, ok := <-asyncQ:
						if !ok {
							return
						}
						b.readq <- data
					default:
					}
				}
			}()
		}
	}
}

func (b *Buffer) Close() {
	b.once.Do(func() {
		b.lock.Lock()
		b.status = RejectStatus
		b.lock.Unlock()

		close(b.sig)
		close(b.bfs[0])
		close(b.bfs[1])
	})
}
