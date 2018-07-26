package connect

import "fmt"

const (
	KILOBYTE = 1024
	MEGABYTE = 1024 * 1024
	GIGABYTE = 1024 * 1024 * 1024
)

type ConnectorStats struct {
	recv int
	send int
}

func (cs *ConnectorStats) Recv() int {
	return cs.recv
}

func (cs *ConnectorStats) Send() int {
	return cs.send
}

func (cs *ConnectorStats) String() string {
	return fmt.Sprintf("recv: %s, send: %s", prettyStats(cs.recv), prettyStats(cs.send))
}

func prettyStats(value int) string {
	switch {
	case value > GIGABYTE:
		return fmt.Sprintf("%d GB", value/GIGABYTE)
	case value > MEGABYTE:
		return fmt.Sprintf("%d MB", value/MEGABYTE)
	case value > KILOBYTE:
		return fmt.Sprintf("%d KB", value/KILOBYTE)
	}
	return fmt.Sprintf("%d Bytes", value)
}
