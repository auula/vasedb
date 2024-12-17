package vfs

import "os"

type Cleaner struct {
	DirtyFiles []*os.File
}
