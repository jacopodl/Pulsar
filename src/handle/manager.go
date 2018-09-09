package handle

import (
	"fmt"
	"strings"
)

const HOPTIONSEP = ":"

var (
	Handlers           = map[string]Handler{}
	hchain   []Handler = nil
)

func Register(handler Handler) error {
	if _, ok := Handlers[handler.Name()]; ok {
		return fmt.Errorf("handle %s already exists", handler.Name())
	}
	Handlers[handler.Name()] = handler
	return nil
}

func MakeChain(hnames []string) error {
	var err error = nil

	for i := range hnames {
		hname, hopts := SplitHandlerOptions(hnames[i])
		if handler, ok := Handlers[hname]; ok {
			if handler, err = handler.Init(hopts); err != nil {
				return err
			}
			hchain = append(hchain, handler)
			continue
		}
		return fmt.Errorf("handle %s not exists", hname)
	}
	return nil
}

func Process(buf []byte, length int, decode bool) ([]byte, int, error) {
	var err error = nil
	for _, handler := range hchain {
		if buf, length, err = handler.Process(buf, length, decode); err != nil || length == 0 {
			buf = nil
			length = 0
			break
		}
	}
	return buf, length, err
}

func SplitHandlerOptions(hopts string) (hname, hoptions string) {
	split := strings.SplitN(hopts, HOPTIONSEP, 2)
	hname = split[0]
	if len(split) > 1 {
		hoptions = split[1]
	}
	return
}
