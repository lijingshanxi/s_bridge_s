package communication

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"s_bridge_s/common"

	"github.com/golang/glog"
)

//CmnctnCtx is
type CmnctnCtx struct {
	EndPoint                string
	RemoteEndPoint          string
	Stop                    bool
	Connected               bool
	PacketChan              chan *common.Packet
	ConnVersion             uint
	Is3389                  bool
	connMutex               *sync.Mutex
	discardPackNum          uint
	discardWriteChanPackNum uint
}

//NewCmnctnCtx is
func NewCmnctnCtx(endPoint string) *CmnctnCtx {
	ret := &CmnctnCtx{
		EndPoint:                endPoint,
		Stop:                    false,
		Connected:               false,
		PacketChan:              nil,
		ConnVersion:             0,
		Is3389:                  false,
		discardPackNum:          0,
		discardWriteChanPackNum: 0,
	}
	ret.Init()
	return ret
}

//Init is
func (c *CmnctnCtx) Init() error {
	c.PacketChan = make(chan *common.Packet, 4096)
	c.connMutex = &sync.Mutex{}
	return nil
}

//Context is
type Context struct {
	Cmnctn1Ctx *CmnctnCtx
	Cmnctn2Ctx *CmnctnCtx
	Wg         *sync.WaitGroup
}

//Communication is
type Communication struct {
	ctx *Context

	conn net.Conn

	peerCommunication *Communication
	myCmnctnCtx       *CmnctnCtx
	peerCmnctnCtx     *CmnctnCtx
}

//Init is
func (c *Communication) Init() error {
	c.peerCommunication = nil
	return nil
}

//SetPeerSocks is
func (c *Communication) SetPeerSocks(peerCommunication *Communication) error {
	c.peerCommunication = peerCommunication
	return nil
}

//SetCtx is
func (c *Communication) SetCtx(ctx *Context) error {
	c.ctx = ctx
	return nil
}

//SetMyCtx is
func (c *Communication) SetMyCtx(sctx *CmnctnCtx) error {
	c.myCmnctnCtx = sctx
	return nil
}

//SetPeerCtx is
func (c *Communication) SetPeerCtx(sctx *CmnctnCtx) error {
	c.peerCmnctnCtx = sctx
	return nil
}

func (c *Communication) closePeer() {
	if c.peerCmnctnCtx.Connected {
		c.peerCommunication.closeConn("BY close peer")
	} else {
		glog.V(10).Infof("[%s]sencondary close peer communication", c.myEndpoint())
	}
}

//recvLoop is
func (c *Communication) recvLoop() {

	for !c.myCmnctnCtx.Stop {
		if !c.myCmnctnCtx.Connected {
			// time.Sleep(time.Duration(100) * time.Millisecond)
			// continue
			break
		}
		bytes, err := c.recvData()
		if err != nil {
			// glog.Warningf("[%s]read err:%s", c.myEndpoint(), err.Error())
			break
		}
		pack := &common.Packet{
			Version: c.myCmnctnCtx.ConnVersion,
			Cmd:     "data",
			Data:    bytes,
		}
		c.writeChan(c.peerCmnctnCtx.PacketChan, pack, "recv loop")
	}

}

func (c *Communication) writeChan(ch chan *common.Packet, pack *common.Packet, by string) error {

	select {
	case ch <- pack:
		{
		}

	case <-time.After(time.Duration(500) * time.Millisecond):
		{
			c.myCmnctnCtx.discardWriteChanPackNum++
			glog.V(10).Infof("[%s]write chan timeout,%s, by:%s", c.myEndpoint(),
				pack.String(), by)
		}
	}

	return nil
}

func (c *Communication) sendData(data []byte) error {
	if !c.myCmnctnCtx.Connected {
		s := fmt.Sprintf("[%s] disconnected, can't send data", c.myCmnctnCtx.EndPoint)
		glog.V(10).Info(s)
		return errors.New(s)
	}
	if c.conn == nil {
		s := fmt.Sprintf("[%s] conn is nil, can't send data", c.myCmnctnCtx.EndPoint)
		glog.V(10).Info(s)
		return errors.New(s)
	}
	if len(data) == 0 {
		glog.Errorf("[%s]error, send length 0", c.myEndpoint())
	}
	// c.conn.SetWriteDeadline(time.Now().Add(1000 * time.Millisecond))
	lastConnVersion := c.myCmnctnCtx.ConnVersion
	_, err := c.conn.Write(data)
	if err != nil {
		// r := c.closeConn("BY Write error")
		// if r == 0 {
		// 	c.closePeer()
		// }
		newConnVersion := c.myCmnctnCtx.ConnVersion
		if lastConnVersion == newConnVersion {
			glog.Errorf("[%s]send data error, %s", c.myEndpoint(), err)
			c.closeConn("BY Write error")
			c.closePeer()
		} else {
			glog.Errorf("[%s]write err, not call close, %d, %d, %s", c.myEndpoint(),
				lastConnVersion, newConnVersion, err.Error())
		}

		return err
	}
	return nil
}

func (c *Communication) recvData() ([]byte, error) {
	buf := make([]byte, 2048)
	// c.conn.SetReadDeadline(time.Now().Add(1000 * time.Millisecond))
	lastConnVersion := c.myCmnctnCtx.ConnVersion
	n, err := c.conn.Read(buf)
	if err != nil {
		// r := c.closeConn("BY recv error")
		// if r == 0 {
		// 	c.closePeer()
		// }
		newConnVersion := c.myCmnctnCtx.ConnVersion
		if lastConnVersion == newConnVersion {
			glog.Errorf("[%s]read err:%s", c.myEndpoint(), err.Error())
			c.closeConn("BY recv error")
			c.closePeer()
		} else {
			glog.Errorf("[%s]read err, not call close, %d, %d, %s", c.myEndpoint(),
				lastConnVersion, newConnVersion, err.Error())
		}

		return nil, err
	}
	if n == 0 {
		glog.Errorf("[%s]error, recv length 0", c.myEndpoint())
	}
	glog.V(20).Infof("[%s]received: length:%d",
		c.myEndpoint(), n)
	return buf[0:n], nil
}

//return: 0, normal; 1:sencondary close
func (c *Communication) closeConn(closeBy string) int {

	c.myCmnctnCtx.connMutex.Lock()
	defer c.myCmnctnCtx.connMutex.Unlock()

	if !c.myCmnctnCtx.Connected {
		glog.V(10).Infof("[%s]sencondary close(%s), by:%s", c.myEndpoint(),
			c.myCmnctnCtx.RemoteEndPoint, closeBy)
		return 1
	}

	c.myCmnctnCtx.Connected = false
	c.myCmnctnCtx.ConnVersion++

	if c.conn != nil {
		glog.V(10).Infof("[%s]close conn(%s), version: %d, by:%s", c.myEndpoint(),
			c.myCmnctnCtx.RemoteEndPoint,
			c.myCmnctnCtx.ConnVersion, closeBy)
		c.conn.Close()
	} else {
		glog.V(10).Infof("[%s]conn is nil(%s), close by:%s", c.myEndpoint(),
			c.myCmnctnCtx.RemoteEndPoint, closeBy)
	}
	return 0
}

func (c *Communication) myEndpoint() string {
	return c.myCmnctnCtx.EndPoint
}
