package types

import "bytes"

type Binary struct {
	buf bytes.Buffer
}

func (bin *Binary) ToBytes() []byte {
	return nil
}
