package handle

import "stringsop"

type stub struct{}

func NewStub() Handler {
	return &stub{}
}

func (s *stub) Name() string {
	return "stub"
}

func (s *stub) Description() string {
	return "Do nothing, pass through"
}

func (s *stub) Init(options []string) error {
	return nil
}

func (s *stub) Options() []stringsop.Option {
	return nil
}

func (s *stub) Process(buf []byte, length int, decode bool) ([]byte, int, error) {
	return buf, length, nil
}
