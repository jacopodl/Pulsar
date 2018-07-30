package connect

import "fmt"

var Connectors = map[string]Connector{}

func RegisterConnector(connector Connector) error {
	if _, ok := Connectors[connector.Name()]; ok {
		return fmt.Errorf("connector %s already exists", connector.Name())
	}
	Connectors[connector.Name()] = connector
	return nil
}

func MakeConnect(cname string, listen bool, address string) (Connector, error) {
	if cnt, err := Connectors[cname]; err {
		return cnt.Connect(listen, address)
	}
	return nil, fmt.Errorf("unknown connector %s, aborted", cname)
}
