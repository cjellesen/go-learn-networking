package internal

type AddrType struct {
	Protocol string
	Addr     string
}

func (a AddrType) Network() string {
	return a.Protocol
}

func (a AddrType) String() string {
	return a.Addr
}

func NewAddrType(protocol string, addr string) AddrType {
	return AddrType{
		Protocol: protocol,
		Addr:     addr,
	}
}
