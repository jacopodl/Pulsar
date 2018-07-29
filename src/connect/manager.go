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
