package communication

import (
	"context"
	"net"
	"s_bridge_s/common"
	"time"

	"github.com/golang/glog"
)

//NewServer is
func NewServer() *Server {
	ret := &Server{}
	ret.Init()
	return ret
}

//Server is
type Server struct {
	Communication
	listener net.Listener
}

//Init is
func (s *Server) Init() {
	s.Communication.Init()
	return
}

//SocketLoop is
func (s *Server) SocketLoop() {
	wg := s.ctx.Wg
	defer wg.Done()

	var err error

	lc := net.ListenConfig{KeepAlive: time.Duration(60 * time.Second)}
	glog.V(4).Info("begin listen:", s.myCmnctnCtx.EndPoint)
	s.listener, err = lc.Listen(context.Background(), "tcp", s.myCmnctnCtx.EndPoint)
	// s.listener, err = net.Listen("tcp", s.myCmnctnCtx.EndPoint)
	if err != nil {
		glog.Error("listen err:", err)
		s.myCmnctnCtx.Stop = true
		return
	}
	defer s.closeServer()
	glog.V(4).Infof("listen on:%s", s.myCmnctnCtx.EndPoint)

	for !s.myCmnctnCtx.Stop {

		glog.V(4).Infof("[%s]waiting client...", s.myEndpoint())
		err = s.accept()
		if err != nil {
			glog.Error("accept err:", err)
			time.Sleep(time.Duration(100) * time.Millisecond)
			continue
		}
		s.recvLoop()
	}
}

func (s *Server) accept() error {
	var err error

	s.myCmnctnCtx.ConnVersion++

	s.conn, err = s.listener.Accept()
	if err != nil {
		glog.Error("accept err:", err)
		return err
	}
	s.myCmnctnCtx.Connected = true
	s.myCmnctnCtx.RemoteEndPoint = s.conn.RemoteAddr().String()
	glog.V(10).Infof("[%s]new client, version:%d, %s", s.myEndpoint(),
		s.myCmnctnCtx.ConnVersion, s.myCmnctnCtx.RemoteEndPoint)
	return nil
}

//ChanLoop is
func (s *Server) ChanLoop() {
	wg := s.ctx.Wg
	defer wg.Done()

	for !s.myCmnctnCtx.Stop {
		var v *common.Packet
		var ok bool
		ok = true
		select {
		case v, ok = <-s.myCmnctnCtx.PacketChan:
		}
		if !ok {
			glog.Errorf("[%s]read chan error", s.myEndpoint())
			// s.onError()
			// s.closeServer()
			break
		}
		glog.V(20).Infof("[%s]chan receive: %s, length:%d",
			s.myEndpoint(), v.Cmd, len(v.Data))
		if v.Cmd == "connect" {
			continue
		}
		if v.Cmd == "disconnect" {
			s.closeConn("BY CHANNEL")
			continue
		}
		if v.Cmd == "data" {
			if s.peerCmnctnCtx.ConnVersion == v.Version {
				s.sendData(v.Data)
			} else {
				s.myCmnctnCtx.discardPackNum++
				glog.V(10).Infof("[%s]chan recv incorrect data, version,%d,%d", s.myEndpoint(),
					s.peerCmnctnCtx.ConnVersion, v.Version)
			}
			continue
		}
		glog.Errorf("[%s]invalid statement", s.myEndpoint())
		glog.Flush()
		break
	}
}

func (s *Server) closeServer() {
	s.closeConn("BY server close")
	s.closeListener()
}

func (s *Server) closeListener() {
	s.myCmnctnCtx.connMutex.Lock()
	defer s.myCmnctnCtx.connMutex.Unlock()

	s.myCmnctnCtx.Connected = false
	s.myCmnctnCtx.ConnVersion++

	if s.listener != nil {
		glog.V(10).Infof("[%s]close listener", s.myEndpoint())
		s.listener.Close()
	} else {
		glog.V(10).Infof("[%s]listener is nil", s.myEndpoint())
	}
}

func (s *Server) myEndpoint() string {
	return s.myCmnctnCtx.EndPoint
}
