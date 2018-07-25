package handle

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

func (s *stub) Process(buf []byte, decode bool) ([]byte, error) {
	return buf, nil
}
