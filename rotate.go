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
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/robfig/cron/v3"
	"golang.org/x/sync/errgroup"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const Layout = "20060102"

// RotateStrategy 日志轮转策略
type RotateStrategy struct {
	// 日志文件目录
	dir string
	// 日志文件名称
	filename string
	// 时区设置，默认Asia/Shanghai
	location string
	// 保存序列化状态的文件路径
	// 当前文件的递增序列号，比如1,2,3,4，用于日志轮转时因为日志量过大，
	// 同一天出现多个日志文件时加上编号进行区分
	sequenceStat *os.File
	// 当前的日志大小
	currentSize int64
	// 当前的日志日期
	currentDate string
	// 日志轮转的阈值
	threshold int64
	// 是否压缩历史日志文件
	enableCompress bool
	// 压缩级别
	compressLevel CompressLevel
	// 加锁保护
	lock sync.RWMutex
	// 文件句柄
	logout *os.File
	// 原生日志
	lg *log.Logger
	// 单例
	once sync.Once
	// 定时任务
	cr *cron.Cron
}

func NewRotateStrategy(filename string, threshold int64, enableCompress bool, level CompressLevel) (*RotateStrategy, error) {
	dir := filepath.Dir(filename)
	sequenceStat, err := os.OpenFile(fmt.Sprintf("%s/sequence.stat", dir),
		os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			_ = sequenceStat.Close()
		}
	}()

	rs := &RotateStrategy{
		dir:            dir,
		filename:       filepath.Base(filename),
		currentDate:    time.Now().Format(Layout),
		threshold:      threshold,
		sequenceStat:   sequenceStat,
		enableCompress: enableCompress,
		compressLevel:  level,
		lock:           sync.RWMutex{},
		lg:             log.New(os.Stdout, "", log.Ldate|log.Lmicroseconds),
		once:           sync.Once{},
	}

	seq, err := rs.loadSequence()
	if err != nil {
		return nil, err
	}

	var fname string
	if seq == 0 {
		// 初次初始化
		if err = rs.saveSequence(0); err != nil {
			return nil, err
		}
		fname = fmt.Sprintf("%s.%s", filename, time.Now().Format(Layout))
	} else {
		// 重新启动，已存在
		fname = fmt.Sprintf("%s/%s.%s.%d.log", rs.dir, rs.filename, time.Now().Format(Layout), seq)
	}
	logout, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	rs.logout = logout

	// 开启一次日志轮转检查
	err = rs.Rotate()
	if err != nil {
		return nil, err
	}

	return rs, nil
}

func (r *RotateStrategy) SetCurrentSize(size int64) {
	atomic.AddInt64(&r.currentSize, size)
}

// AsyncWork 开启一个异步的定时任务，每天凌晨24点准时进行日志轮转，定时任务精确到秒，生成新一天的日志文件
func (r *RotateStrategy) AsyncWork() {
	r.once.Do(func() {
		location, err := time.LoadLocation(r.location)
		if err != nil {
			_, _ = os.Stderr.WriteString("load location fail:" + err.Error())
			return
		}

		r.cr = cron.New(
			cron.WithLocation(location),
			cron.WithSeconds())
		entity, err := r.cr.AddFunc("0 0 0 * * *", func() {
			r.lock.Lock()
			defer r.lock.Unlock()

			_ = r.logout.Close()
			srcFileName := r.logout.Name()
			if r.enableCompress {
				var eg errgroup.Group
				eg.Go(func() error {
					return r.compress(srcFileName)
				})
				if err := eg.Wait(); err == nil {
					_ = os.Remove(srcFileName)
				}
			}

			logout, err := os.OpenFile(fmt.Sprintf("%s/%s.%s", r.dir, r.filename, time.Now().Format(Layout)),
				os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				_, _ = os.Stderr.WriteString(fmt.Sprintf("failed to open filename: %s, err: %v", r.filename, err))
				return
			}
			r.logout = logout
			r.lg = log.New(logout, "", log.Ldate|log.Lmicroseconds)
			r.currentDate = time.Now().Format(Layout)
			atomic.StoreInt64(&r.currentSize, 0)

			_, err = r.sequenceStat.WriteString("0")
			if err != nil {
				_, _ = os.Stderr.WriteString(fmt.Sprintf("failed to set sequence stat, err: %v", err))
				return
			}
		})

		if err != nil {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("failed to add rotate cron job, err: %v", err))
			return
		}

		_, _ = os.Stdin.WriteString(fmt.Sprintf("add rotate cron job, entity: %d \n", entity))
		r.cr.Start()
	})
}

// Rotate 日志轮转的实现方法，轮转逻辑如下：
// 1. 每次写入时检查当前日期是否已经距离当前日志文件创建时相差一天，如果是则进行轮转
// 2. 每次写入时检查当前日志文件大小是否已经达到文件的轮转阈值，如果时则进行轮转
func (r *RotateStrategy) Rotate() error {
	r.lock.RLock()
	date := time.Now().Format(Layout)
	// 快路径
	if date == r.currentDate && r.currentSize < r.threshold {
		r.lock.RUnlock()
		return nil
	}
	r.lock.RUnlock()

	r.lock.Lock()
	defer r.lock.Unlock()
	srcFileName := r.logout.Name()

	if date != r.currentDate {
		_ = r.logout.Close()
		if r.enableCompress {
			var eg errgroup.Group
			eg.Go(func() error {
				return r.compress(srcFileName)
			})
			if err := eg.Wait(); err == nil {
				_ = os.Remove(srcFileName)
			}
		}

		if err := r.createNewFile(r.filename, 0); err != nil {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("failed to open new file, filename: %s, err: %v", r.filename, err))
			return err
		}
		return nil
	}

	if r.currentSize >= r.threshold {
		_ = r.logout.Close()
		if r.enableCompress {
			var eg errgroup.Group
			eg.Go(func() error {
				return r.compress(srcFileName)
			})
			if err := eg.Wait(); err == nil {
				_ = os.Remove(srcFileName)
			}
		}

		seq, err := r.loadSequence()
		if err != nil {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("failed to load sequence, err: %v", err))
			return err
		}
		newSeq := seq + 1

		fileName := fmt.Sprintf("%s.%s.%d.log", r.filename, date, newSeq)
		err = r.createNewFile(fileName, newSeq)
		if err != nil {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("failed to create new file, filename: %s, err: %v", fileName, err))
			return err
		}
	}

	return nil
}

// 读取序列号时重置文件指针
func (r *RotateStrategy) loadSequence() (int, error) {
	_, err := r.sequenceStat.Seek(0, 0)
	if err != nil {
		return 0, err
	}

	data, err := io.ReadAll(r.sequenceStat)
	if err != nil {
		return 0, err
	}

	if len(data) == 0 {
		return 0, nil
	}

	seq, err := strconv.Atoi(string(bytes.TrimSpace(data)))
	if err != nil {
		return 0, err
	}
	return seq, nil
}

// 写入序列号时清空文件
func (r *RotateStrategy) saveSequence(seq int) error {
	err := r.sequenceStat.Truncate(0)
	if err != nil {
		return err
	}

	_, err = r.sequenceStat.Seek(0, 0)
	if err != nil {
		return err
	}
	_, err = r.sequenceStat.WriteString(strconv.Itoa(seq))
	return err
}

func (r *RotateStrategy) createNewFile(filename string, seq int) error {
	logout, err := os.OpenFile(fmt.Sprintf("%s/%s", r.dir, filename), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	r.lg = log.New(logout, "", log.Ldate|log.Lmicroseconds)
	r.logout = logout
	atomic.StoreInt64(&r.currentSize, 0)

	return r.saveSequence(seq)
}

// compress 执行压缩操作
func (r *RotateStrategy) compress(srcFilename string) error {
	srcFile, err := os.OpenFile(srcFilename, os.O_RDONLY, 0666)
	if err != nil {
		_, _ = os.Stderr.WriteString(fmt.Sprintf("failed to open gzip file, filename: %s, err: %v", r.filename, err))
		return err
	}
	defer func() {
		_ = srcFile.Close()
	}()

	gzFile, err := os.OpenFile(fmt.Sprintf("%s.gz", srcFilename),
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		_, _ = os.Stderr.WriteString(fmt.Sprintf("failed to open gzip file, filename: %s, err: %v", r.filename, err))
		return err
	}
	defer func() {
		_ = gzFile.Close()
	}()

	gzWriter, err := gzip.NewWriterLevel(gzFile, r.compressLevel.Int())
	if err != nil {
		_, _ = os.Stderr.WriteString(fmt.Sprintf("failed to open gzip file, filename: %s, err: %v", r.filename, err))
		return err
	}
	defer func() {
		_ = gzWriter.Close()
	}()

	// 每次读取1M的内容
	buf := make([]byte, 1024*1024)
	for {
		n, err := srcFile.Read(buf)
		if err != nil && err != io.EOF {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("failed to read src file to gzip file, err: %v", err))
			return err
		}

		if n == 0 {
			break
		}

		if _, err = gzWriter.Write(buf[:n]); err != nil {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("failed to write src file to gzip file, err: %v", err))
			return err
		}
	}
	return gzWriter.Flush()
}

// Close 关闭轮转功能
func (r *RotateStrategy) Close() {
	r.cr.Stop()
	_ = r.logout.Close()
	_ = r.sequenceStat.Close()
}
