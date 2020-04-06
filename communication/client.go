package communication

import (
	"net"
	"s_bridge_s/common"
	"time"

	"github.com/golang/glog"
)

//NewClient is
func NewClient() *Client {
	ret := &Client{}
	ret.Init()
	return ret
}

//Client is
type Client struct {
	Communication
}

//Init is
func (c *Client) Init() {
	c.Communication.Init()
	return
}

//SocketLoop is
func (c *Client) SocketLoop() {
	wg := c.ctx.Wg
	defer wg.Done()

	var err error

	for !c.myCmnctnCtx.Stop {
		if !c.myCmnctnCtx.IsClientTo3389 {
			err = c.connect()
			if err != nil {
				// time.Sleep(time.Duration(100) * time.Millisecond)
				continue
			}
		}
		c.recvLoop()
	}
}

func (c *Client) connect() error {
	if c.myCmnctnCtx.Connected {
		return nil
	}

	var err error

	c.myCmnctnCtx.ConnVersion++

	d := net.Dialer{
		KeepAlive: time.Duration(60 * time.Second),
		// Timeout:   10 * time.Second,
	}
	c.conn, err = d.Dial("tcp", c.myCmnctnCtx.EndPoint)
	// c.conn, err = net.Dial("tcp", c.myCmnctnCtx.EndPoint)
	if err != nil {
		glog.V(21).Info("Dial err:", err)
		return err
	}
	c.myCmnctnCtx.Connected = true
	c.myCmnctnCtx.RemoteEndPoint = c.conn.RemoteAddr().String()
	glog.V(4).Infof("[%s]connected, version:%d, %s", c.myEndpoint(),
		c.myCmnctnCtx.ConnVersion, c.myCmnctnCtx.EndPoint)
	return nil
}

//ChanLoop is
func (c *Client) ChanLoop() {
	wg := c.ctx.Wg
	defer wg.Done()

	for !c.myCmnctnCtx.Stop {
		var v *common.Packet
		var ok bool
		ok = true
		select {
		case v, ok = <-c.myCmnctnCtx.PacketChan:
		}
		if !ok {
			glog.Errorf("[%s]read chan error", c.myEndpoint())
			break
		}
		glog.V(20).Infof("[%s]chan receive: %s, length:%d",
			c.myEndpoint(), v.Cmd, len(v.Data))
		if v.Cmd == "data" {
			if c.peerCmnctnCtx.ConnVersion == v.Version {
				if c.myCmnctnCtx.IsClientTo3389 {
					c.connect()
				}
				c.sendData(v.Data)
			} else {
				c.myCmnctnCtx.discardPackNum++
				glog.V(10).Infof("[%s]chan recv incorrect data, version,%d,%d", c.myEndpoint(),
					c.peerCmnctnCtx.ConnVersion, v.Version)
			}
			continue
		}
		glog.Errorf("[%s]invalid statement", c.myEndpoint())
		glog.Flush()
		break
	}
}

func (c *Client) myEndpoint() string {
	return c.myCmnctnCtx.EndPoint
}
