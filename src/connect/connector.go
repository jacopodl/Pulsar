package connect

type Connector interface {
	Name() string
	Description() string
	Stats() *ConnectorStats
	Connect(listen bool, address string) (Connector, error)
	Close()
	Read() ([]byte, int, error)
	Write(buf []byte, length int) (int, error)
}
