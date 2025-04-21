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

import "compress/gzip"

type CompressLevel int

const (
	NoCompression      CompressLevel = gzip.NoCompression
	BestSpeed          CompressLevel = gzip.BestSpeed
	BestCompression                  = gzip.BestCompression
	DefaultCompression               = gzip.DefaultCompression
	HuffmanOnly                      = gzip.HuffmanOnly
)

func (l CompressLevel) valid() bool {
	switch l {
	case BestSpeed, BestCompression, DefaultCompression, HuffmanOnly:
		return true
	default:
		return false
	}
}

func (l CompressLevel) Int() int {
	return int(l)
}
