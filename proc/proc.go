package proc

import (
	"s_bridge_s/communication"
	"sync"
)

//RunSS is
func RunSS(server1, server2 string) {

	var wg sync.WaitGroup
	s1Ctx := communication.NewCmnctnCtx(server1)
	s2Ctx := communication.NewCmnctnCtx(server2)
	ctx := &communication.Context{
		Cmnctn1Ctx: s1Ctx,
		Cmnctn2Ctx: s2Ctx,
		Wg:         &wg,
	}
	s1 := communication.NewServer()
	s1.SetCtx(ctx)
	s1.SetMyCtx(s1Ctx)
	s1.SetPeerCtx(s2Ctx)

	s2 := communication.NewServer()
	s2.SetCtx(ctx)
	s2.SetMyCtx(s2Ctx)
	s2.SetPeerCtx(s1Ctx)

	s1.SetPeerSocks(&s2.Communication)
	s2.SetPeerSocks(&s1.Communication)

	wg.Add(1)
	go s1.SocketLoop()
	wg.Add(1)
	go s1.ChanLoop()

	wg.Add(1)
	go s2.SocketLoop()
	wg.Add(1)
	go s2.ChanLoop()

	wg.Wait()
}

//RunCC is
func RunCC(server1, server2 string) {

	var wg sync.WaitGroup
	s1Ctx := communication.NewCmnctnCtx(server1)
	//We assume that the first IP is to connect to 3389
	s1Ctx.Is3389 = true
	s2Ctx := communication.NewCmnctnCtx(server2)
	ctx := &communication.Context{
		Cmnctn1Ctx: s1Ctx,
		Cmnctn2Ctx: s2Ctx,
		Wg:         &wg,
	}
	s1 := communication.NewClient()
	s1.SetCtx(ctx)
	s1.SetMyCtx(s1Ctx)
	s1.SetPeerCtx(s2Ctx)
	s2 := communication.NewClient()
	s2.SetCtx(ctx)
	s2.SetMyCtx(s2Ctx)
	s2.SetPeerCtx(s1Ctx)

	s1.SetPeerSocks(&s2.Communication)
	s2.SetPeerSocks(&s1.Communication)

	wg.Add(1)
	go s1.SocketLoop()
	wg.Add(1)
	go s1.ChanLoop()

	wg.Add(1)
	go s2.SocketLoop()
	wg.Add(1)
	go s2.ChanLoop()

	wg.Wait()

}
