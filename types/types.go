package types

type QueryRow struct {
	Key   string `json:"key"`
	Index int    `json:"index,omitempty"`
	Range string `json:"range,omitempty"`
	Order string `json:"order,omitempty"`
	Score string `json:"score,omitempty"`
	Field string `json:"field,omitempty"`
}

type Query struct {
	Type  string  `json:"type"`
	Query []Query `json:"query"`
}

type Queryer interface {
	Search(qs []Query) []byte
}

type StrQuery struct {
	Query
}

func (sq *StrQuery) Search(qs []Query) []byte {
	return []byte{}
}

type HashQuery struct {
	Query
}

type ListQuery struct {
	Query
}

type SetQuery struct {
	Query
}

type ZSetQuery struct {
	Query
}
