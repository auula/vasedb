// Open Source: MIT License
// Author: Leon Ding <ding@ibyte.me>
// Date: 2022/2/27 - 4:26 下午 - UTC/GMT+08:00

package bottle

import (
	"testing"
	"time"
)

func TestBinaryEncode(t *testing.T) {
	item := NewItem([]byte("foo"), []byte("bar"), uint64(time.Now().Unix()))
	if cap(BinaryEncode(item)) != len(BinaryEncode(item)) {
		t.Error("binary encode buffer size discord")
	}
}
