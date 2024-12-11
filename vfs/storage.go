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

type Storage struct {
	KeyDir      map[string]Index
	ActiveFile  *DataFile
	ArchiveFile map[string]*DataFile
}

type Index struct {
}

func NewStorage() *Storage {
	return nil
}

func (s *Storage) Put() error {
	return nil
}

func (s *Storage) Get() error {
	return nil
}
