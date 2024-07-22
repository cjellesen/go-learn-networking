package internal

type AddrType struct {
	protocol string
	addr     string
}

func NewAddrType(protocol string, addr string) AddrType {
	return AddrType{
		protocol: protocol,
		addr:     addr,
	}
}

func (a *AddrType) Network() string {
	return a.protocol
}

func (a *AddrType) String() string {
	return a.addr
}
