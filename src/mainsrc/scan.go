package main

import (
	"graph"
	"log"
	_ "mock"
	"os"
	. "scanner"
	"sync"
	"util"
)

func SaveNetNodes(netgraph *graph.NetGraph, netnodes []*util.NetNode) (map[string]int64, error) {

	err := netgraph.DropDatabase()
	if err != nil {
		return nil, err
	}

	_ = netgraph.DropIndexOnNetNodeID()

	err = netgraph.TxStart()
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

	err = netgraph.CreateIndexOnNetNodeID()
	if err != nil {
		return nil, err
	}

	return nodeids, nil
}
func main() {

	const (
		MaxNetChassisIdChanNum    = 10000
		MaxNetChassisIdNum        = 30000
		MaxUnValidNeighborChanNum = 20000
		MaxValidNeighborChanNum   = 10000
		Community                 = "360buy"
	)

	//the config
	const configfile = "./config.json"
	config, err := util.NewConfig(configfile)
	if err != nil {
		util.Logger.Printf("Failed to get configuration infomations. %v\n", err)
		os.Exit(1)
	}

	// the batch commit number
	var CommitBatch = config.SaveBatch

	//Init logging file.
	logFile := config.LogFile
	logbufer, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 666)
	if err != nil {
		log.Printf("Failed to Open File. %v\n", err)
	}
	defer logbufer.Close()

	util.Logger = log.New(logbufer, "[INFO]", log.LstdFlags)

	netnodes, err := GetNetNode(config.Url)
	if err != nil {
		util.Logger.Printf("Failed to get netnode infomations. %v\n", err)
		os.Exit(1)
	}

	boltserver := config.NeoServer
	boltname := config.NeoUser
	boltpwd := config.NeoPassword
	netgraph := graph.NewNetGraph(boltserver, boltname, boltpwd, 0)

	err = netgraph.ConnectNeo4j()
	if err != nil {
		util.Logger.Printf("Connect Neo4j Server Failed. %v\n", err)
		os.Exit(1)
	}
	defer netgraph.Exit()

	nodeids, err := SaveNetNodes(netgraph, netnodes)
	if err != nil {
		util.Logger.Printf("Save Nodes Failed. %v\n")
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
		SavedCount:          0,
	}

	worker.SaveFinished.Add(1)

	go worker.GenerateNeighbor()

	worker.ReadChannel()

	err = netgraph.TxStart()
	if err != nil {
		util.Logger.Printf("%v\n", err)
		os.Exit(1)
	}

	saveneighbor := func(neighbor *NetNeighbor) error {
		err := netgraph.CreateNetLinkByNetNodeIDWithTX(
			nodeids[neighbor.LocalIP],
			nodeids[neighbor.RemoteIP],
			neighbor.LocalPort,
			neighbor.RemotePort)
		if worker.SavedCount > CommitBatch {
			log.Println("StartCommit")
			err = netgraph.TxCommit()
			if err != nil {
				_ = netgraph.TxRollback()
				return err
			}

			worker.SavedCount = 0

			err = netgraph.TxClose()
			if err != nil {
				_ = netgraph.TxRollback()
				return err

			}
			err = netgraph.TxStart()
			if err != nil {
				return err
			}
			worker.SavedCount = 0
			log.Println("FinishedCommit")
		}
		return err
	}

	worker.SafeSaveNeighbor(saveneighbor)

	worker.SaveFinished.Wait()

	err = netgraph.TxCommit()
	if err != nil {
		_ = netgraph.TxRollback()
		util.Logger.Printf("Save Neighbor TxCommit Failed, Rollbacked. %v\n", err)
	}

	err = netgraph.TxClose()
	if err != nil {
		_ = netgraph.TxRollback()
		util.Logger.Printf("Save Neighbor TxClose Failed, Rollbacked. %v\n", err)

	}

	util.Logger.Printf("Scan Completed!")
}
