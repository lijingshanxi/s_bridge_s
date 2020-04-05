package common

import "fmt"

//Packet is
type Packet struct {
	Version uint
	Cmd     string
	Data    []byte
}

//String is
func (p *Packet) String() string {
	ret := fmt.Sprintf("Cmd:%s,length:%d,ver:%d", p.Cmd, len(p.Data), p.Version)
	return ret
}
