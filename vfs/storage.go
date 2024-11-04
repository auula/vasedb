package vfs

type Entry struct {
	key   []byte
	value []byte
}

func NewEntry(key, value string) *Entry {
	return &Entry{
		key:   []byte(key),
		value: []byte(value),
	}
}

func (e *Entry) Key() []byte {
	return e.key
}

func (e *Entry) Value() []byte {
	return e.value
}

type Storage interface {
	BatchPut(keys []*Entry) error
	Get(key *Entry) error
	Put(key *Entry) error
}

type BitCask struct {
	KeyDir      map[string]interface{}
	ActiveFile  *FileSystem
	ArchiveFile []*FileSystem
}

func NewBitCask() *BitCask {
	return nil
}
