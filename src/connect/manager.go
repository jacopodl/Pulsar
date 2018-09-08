package connect

import (
	"fmt"
	"strings"
)

const COPTIONSEP = ":"

var Connectors = map[string]Connector{}

func Register(connector Connector) error {
	if _, ok := Connectors[connector.Name()]; ok {
		return fmt.Errorf("connector %s already exists", connector.Name())
	}
	Connectors[connector.Name()] = connector
	return nil
}

func MakeConnect(cnameopt string, listen, plain bool) (Connector, error) {
	connector, address := parseConnectorOptions(cnameopt)
	if cnt, err := Connectors[connector]; err {
		return cnt.Connect(listen, plain, address)
	}
	return nil, fmt.Errorf("unknown connector %s, aborted", connector)
}

func parseConnectorOptions(value string) (connector string, address string) {
	tmp := strings.SplitN(value, COPTIONSEP, 2)
	connector = tmp[0]
	if len(tmp) > 1 {
		address = tmp[1]
	}
	return
}
