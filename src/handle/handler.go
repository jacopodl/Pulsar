package handle

type Handler interface {
	Name() string
	Description() string
	Process(buf []byte, length int, decode bool) ([]byte, int, error)
}
