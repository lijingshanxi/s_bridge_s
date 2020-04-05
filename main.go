package main

import (
	"flag"
	"fmt"
	"s_bridge_s/proc"

	"github.com/golang/glog"
)

func usage() {
	fmt.Println("usage...")

}

//-gcflags '-N -l'
func main() {

	fmt.Println("start")

	var runType, server1, server2 string
	flag.StringVar(&runType, "st", "type is null", "-st type")
	flag.StringVar(&server1, "s1", "server 1 is null", "-s1 ip")
	flag.StringVar(&server2, "s2", "server 2 is null", "-s2 ip")
	flag.Parse()
	if runType == "" {
		usage()
		s := "runType is null"
		glog.Error(s)
		return
	}
	if server1 == "" {
		usage()
		s := "server1 ip is null"
		glog.Error(s)
		return
	}
	if server1 == "" {
		usage()
		s := "server1 ip is null"
		glog.Error(s)
		return
	}
	if server2 == "" {
		usage()
		s := "server2 ip is null"
		glog.Error(s)
		return
	}

	if runType == "ss" {
		proc.RunSS(server1, server2)
	}
	if runType == "cc" {
		proc.RunCC(server1, server2)
	}

	glog.V(3).Infof("end...")

}
