package main

import (
	"fmt"
	. "mock"
	"sync"
	"time"
)

func main() {
	const URL = "http://api.joybase.jd.com/network_devices?management_ip=172.28.1.1,172.28.1.5&service_status=%E5%9C%A8%E7%BA%BF"
	worker := &NetNeighborScanner{
		NetnodeChan:         make(chan *NetNode, 100),
		NetChassisIdChan:    make(chan [2]string, 100),
		NetChassisId:        NewSafeMap(300),
		UnValidNeighborChan: make(chan *NetNeighbor, 100),
		UnValidNeighbor:     NodeListInit(),
		ValidNeighborChan:   make(chan *NetNeighbor, 100),
		Community:           "360buy",
		GetNodeFinished:     false,
		ScanFinished:        false,
		SaveFinished:        sync.WaitGroup{},
	}
	worker.SaveFinished.Add(1)

	go func() {
		if err := worker.GetNetNodeMock(URL); err != nil {
			fmt.Println(err)
		}
	}()

	go worker.GenerateNeighbor()
	worker.ReadChannel()
	worker.SaveNeighbor()
	worker.SaveFinished.Wait()
	time.Sleep(1 * time.Second)
	fmt.Println("Scan Completed!")

}
