package handle

import "stringsop"

type Handler interface {
	Name() string
	Description() string
	Init(options []string) error
	Options() []stringsop.Option
	Process(buf []byte, length int, decode bool) ([]byte, int, error)
}
