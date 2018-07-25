package handle

type Handler interface {
	Name() string
	Description() string
	Process(buf []byte, decode bool) ([]byte, error)
}
