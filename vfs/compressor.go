package vfs

import "os"

type Compressor struct {
	DirtyReginos []*os.File
}
