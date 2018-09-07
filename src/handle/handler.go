package handle

type Handler interface {
	Name() string
	Description() string
	Init(options string) (Handler, error)
	Process(buf []byte, length int, decode bool) ([]byte, int, error)
}
