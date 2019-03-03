package main

import (
	"fmt"
	"graph"
	. "mock"
	"sync"
	"time"
)

func run(url string) {
	netnodes, err := GetNetNodeMock(url)
	if err != nil {
		fmt.Println(err)
	}
	netgraph := graph.NewNetGraph("bolt://nw.jd.com:80", "neo4j", "wearenetwork", 0)

	err = netgraph.ConnectNeo4j()
	if err != nil {
		fmt.Printf("%v", err)
	}
	defer netgraph.Exit()

	err = netgraph.TxStart()
	if err != nil {
		fmt.Printf("%v", err)
	}

	//Store the node id.
	nodeids := map[string]int64{}
	for _, node := range netnodes {
		nodeids[node.Mgt] = node.Id
		err = netgraph.CreateNetNodeWithTx(node)
		if err != nil {
			fmt.Printf("%v", err)
		}
	}

	err = netgraph.TxCommit()
	if err != nil {
		fmt.Printf("%v", err)
	}

	const (
		MaxNetChassisIdChanNum    = 100
		MaxNetChassisIdNum        = 300
		MaxUnValidNeighborChanNum = 100
		MaxValidNeighborChanNum   = 100
		Community                 = "360buy"
	)

	saveneighbor := func(neighbor *NetNeighbor) error {
		err := netgraph.CreateNetLinkByNetNodeID(
			nodeids[neighbor.LocalIP],
			nodeids[neighbor.RemoteIP],
			neighbor.LocalPort,
			neighbor.RemotePort)
		if err != nil {
			fmt.Printf("%v", err)
		}
		return err
	}

	worker := &NetNeighborScanner{
		NetNodes:            netnodes,
		NetChassisIdChan:    make(chan [2]string, MaxNetChassisIdChanNum),
		NetChassisId:        NewSafeMap(MaxNetChassisIdNum),
		UnValidNeighborChan: make(chan *NetNeighbor, MaxUnValidNeighborChanNum),
		UnValidNeighbor:     NodeListInit(),
		ValidNeighborChan:   make(chan *NetNeighbor, MaxValidNeighborChanNum),
		Community:           Community,
		ScanFinished:        false,
		SaveFinished:        sync.WaitGroup{},
	}

	worker.SaveFinished.Add(1)

	go worker.GenerateNeighbor()

	worker.ReadChannel()

	worker.SaveNeighbor(saveneighbor)

	worker.SaveFinished.Wait()

	time.Sleep(1 * time.Second)
	fmt.Println("Scan Completed!")
}

func main() {
	const URL = "http://api.joybase.jd.com/network_devices?management_ip=172.28.1.1,172.28.1.5&service_status=%E5%9C%A8%E7%BA%BF"
	run(URL)
	/*
		netnodes, err := GetNetNodeMock(URL)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(netnodes)


		worker := &NetNeighborScanner{
			NetNodes:         	 netnodes,
			NetChassisIdChan:    make(chan [2]string, 100),
			NetChassisId:        NewSafeMap(300),
			UnValidNeighborChan: make(chan *NetNeighbor, 100),
			UnValidNeighbor:     NodeListInit(),
			ValidNeighborChan:   make(chan *NetNeighbor, 100),
			Community:           "360buy",
			ScanFinished:        false,
			SaveFinished:        sync.WaitGroup{},
		}

		worker.SaveFinished.Add(1)

		go worker.GenerateNeighbor()
		worker.ReadChannel()
		worker.SaveNeighbor(saveneighbor)
		worker.SaveFinished.Wait()
		time.Sleep(1 * time.Second)
		fmt.Println("Scan Completed!") */

}
