package vm

// Codes source code text
type Codes []byte

// Engine code execution engine
type Engine interface {
	Execute(Codes) ([]byte, error)
}

// Wasm is webassebly engine
type Wasm struct{}

func (ws *Wasm) Execute(code Codes) ([]byte, error) {
	return nil, nil
}

// ECMAScript is javascript engine
type ECMAScript struct{}

func (es *ECMAScript) Execute(code Codes) ([]byte, error) {
	return nil, nil
}
