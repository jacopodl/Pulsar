package handle

import (
	"fmt"
)

var (
	Handlers            = map[string]Handler{}
	Hchain   []*Handler = nil
)

func RegisterHandler(handler Handler) error {
	if _, ok := Handlers[handler.Name()]; ok {
		return fmt.Errorf("handle %s already exists", handler.Name())
	}
	Handlers[handler.Name()] = handler
	return nil
}

func MakeChain(hnames []string, options []string) error {
	for _, hname := range hnames {
		if handler, ok := Handlers[hname]; ok {
			if err := handler.Init(options); err != nil {
				return err
			}
			Hchain = append(Hchain, &handler)
			continue
		}
		return fmt.Errorf("handle %s not exists", hname)
	}
	return nil
}

func Process(buf []byte, length int, decode bool) ([]byte, int, error) {
	var err error = nil
	for _, handler := range Hchain {
		if buf, length, err = (*handler).Process(buf, length, decode); err != nil {
			return nil, 0, err
		}
	}
	return buf, length, err
}
