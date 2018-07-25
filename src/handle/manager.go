package handle

import (
	"fmt"
)

var (
	Handlers = map[string]Handler{}
)

func AddHandler(handler Handler) error {
	if _, ok := Handlers[handler.Name()]; ok {
		return fmt.Errorf("handle %s already exists", handler.Name())
	}
	Handlers[handler.Name()] = handler
	return nil
}
