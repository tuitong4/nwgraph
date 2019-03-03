package main

import (
	"graph"
	"log"
	. "mock"
	"os"
	"sync"
	"util"
)

func SaveNetNodes(netgraph *graph.NetGraph, netnodes []*util.NetNode) (map[string]int64, error) {
	err := netgraph.TxStart()
	if err != nil {
		return nil, err
	}

	//Store the node id.
	nodeids := map[string]int64{}
	for _, node := range netnodes {
		nodeids[node.Mgt] = node.Id
		err = netgraph.CreateNetNodeWithTx(node)
		if err != nil {
			return nil, err
		}
	}

	err = netgraph.TxCommit()
	if err != nil {
		_ = netgraph.TxRollback()
		return nil, err
	}

	err = netgraph.TxClose()
	if err != nil {
		_ = netgraph.TxRollback()
		return nil, err
	}

	return nodeids, nil
}
func main() {

	const (
		MaxNetChassisIdChanNum    = 100
		MaxNetChassisIdNum        = 300
		MaxUnValidNeighborChanNum = 100
		MaxValidNeighborChanNum   = 100
		Community                 = "360buy"
	)

	const configfile = "./config.json"
	config, err := util.NewConfig(configfile)
	if err != nil {
		log.Printf("Failed to get configuration infomations. %v", err)
		os.Exit(1)
	}

	netnodes, err := GetNetNodeMock(config.Url)
	if err != nil {
		log.Printf("Failed to get netnode infomations. %v", err)
		os.Exit(1)
	}

	netgraph := graph.NewNetGraph("bolt://nw.jd.com:443", "neo4j", "wearenetwork", 0)

	err = netgraph.ConnectNeo4j()
	if err != nil {
		log.Printf("Connect Neo4j Server Failed. %v", err)
		os.Exit(1)
	}
	defer netgraph.Exit()

	nodeids, err := SaveNetNodes(netgraph, netnodes)
	if err != nil {
		log.Printf("Save Nodes Failed. %v")
		os.Exit(1)
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

	err = netgraph.TxStart()
	if err != nil {
		log.Printf("%v", err)
		os.Exit(1)
	}

	saveneighbor := func(neighbor *NetNeighbor) error {
		err := netgraph.CreateNetLinkByNetNodeIDWithTX(
			nodeids[neighbor.LocalIP],
			nodeids[neighbor.RemoteIP],
			neighbor.LocalPort,
			neighbor.RemotePort)
		return err
	}

	worker.SaveNeighbor(saveneighbor)

	worker.SaveFinished.Wait()

	err = netgraph.TxCommit()
	if err != nil {
		_ = netgraph.TxRollback()
		log.Printf("Save Neighbor TxCommit Failed, Rollbacked. %v", err)
	}

	err = netgraph.TxClose()
	if err != nil {
		_ = netgraph.TxRollback()
		log.Printf("Save Neighbor TxClose Failed, Rollbacked. %v", err)

	}

	log.Printf("Scan Completed!")
}
