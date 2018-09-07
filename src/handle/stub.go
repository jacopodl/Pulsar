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

func (s *stub) Init(options string) (Handler, error) {
	return s, nil
}

func (s *stub) Process(buf []byte, length int, decode bool) ([]byte, int, error) {
	return buf, length, nil
}
