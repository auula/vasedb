package vfs

import (
	"fmt"

	"github.com/auula/vasedb/types"
)

type Kind int8

const (
	Set Kind = iota
	ZSet
	List
	Text
	Tables
	Binary
	Number
)

type Segment struct {
	kind Kind
	data []byte
}

type Serializable interface {
	ToBytes() []byte
}

// NewSegment 使用数据类型初始化并返回对应的 Segment
func NewSegment(data Serializable) (*Segment, error) {
	var kind Kind
	switch data.(type) {
	case *types.Set:
		kind = Set
	case *types.ZSet:
		kind = ZSet
	case *types.List:
		kind = List
	case *types.Text:
		kind = Text
	case *types.Tables:
		kind = Tables
	case *types.Binary:
		kind = Binary
	case *types.Number:
		kind = Number
	default:
		// 如果类型不匹配，则返回 nil
		return nil, fmt.Errorf("unsupported data type: %T", data)
	}

	// 如果类型不匹配，则返回错误
	return &Segment{
		kind: kind,
		data: data.ToBytes(),
	}, nil
}

func (s *Segment) Kind() Kind {
	return s.kind
}

func (s *Segment) Size() int {
	return len(s.data)
}

func (s *Segment) ToBytes() []byte {

	return []byte{}
}

func (s *Segment) ToSet() *types.Set {
	if s.kind != Set {
		return nil
	}
	// 假设您的数据是 JSON 或某种结构体，可以进行反序列化
	var set types.Set
	// Deserialize s.data to set (具体根据类型定义来做反序列化)
	return &set
}

func (s *Segment) ToZSet() *types.ZSet {
	return nil
}

func (s *Segment) ToText() *types.Text {
	return nil
}

func (s *Segment) ToList() *types.List {
	return nil
}

func (s *Segment) ToTables() *types.Tables {
	return nil
}

func (s *Segment) ToBinary() *types.Binary {
	return nil
}

func (s *Segment) ToNumber() *types.Number {
	return nil
}
