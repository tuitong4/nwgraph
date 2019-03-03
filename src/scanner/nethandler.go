package scanner

import (
	"fmt"
	"sync"
	"time"
	. "util"
)

type NetNeighbor struct {
	LocalIP    string
	LocalPort  []string
	RemoteIP   string
	RemotePort []string
}

type NetNeighborScanner struct {
	NetNodes            []*NetNode
	NetChassisIdChan    chan [2]string
	NetChassisId        SafeMap
	UnValidNeighborChan chan *NetNeighbor
	UnValidNeighbor     *NodeList
	ValidNeighborChan   chan *NetNeighbor
	Community           string
	ScanFinished        bool
	SaveFinished        sync.WaitGroup
}

func (n *NetNeighborScanner) scanNeighbor(netnode *NetNode) error {
	var nodehandler *NetNodeHandler
	nodehandler = NewNetNodeHandler(netnode, n.Community)
	if err := nodehandler.SNMPConnect(); err != nil {
		return err
	}
	defer nodehandler.SNMPClose()

	self_chassis, err := nodehandler.SelfChassisID()
	if err != nil {
		return err
	}
	for _, id := range self_chassis {
		n.NetChassisIdChan <- [2]string{id, nodehandler.node.Mgt}
	}

	rem_chassis, err := nodehandler.RemChassisID()
	if err != nil {
		return err
	}

	rem_port, err := nodehandler.RemPort()
	if err != nil {
		return err
	}

	local_port, err := nodehandler.LocalPort()
	if err != nil {
		return err
	}

	neighbors := map[string]*NetNeighbor{}

	for rem_idx, chassis := range rem_chassis {
		if _, ok := neighbors[chassis]; !ok {
			neighbors[chassis] = &NetNeighbor{
				LocalIP:    nodehandler.node.Mgt,
				LocalPort:  []string{},
				RemoteIP:   "",
				RemotePort: []string{},
			}
		}
		if localportname, ok := local_port[rem_idx]; ok {
			neighbors[chassis].LocalPort = append(neighbors[chassis].LocalPort, localportname)
			neighbors[chassis].RemotePort = append(neighbors[chassis].RemotePort, rem_port[rem_idx])
		}
	}

	for chassis, neighbor := range neighbors {
		if rem_ip, ok := n.NetChassisId.Get(chassis); ok {
			neighbor.RemoteIP = rem_ip
			n.ValidNeighborChan <- neighbor
		} else {
			neighbor.RemoteIP = chassis
			n.UnValidNeighborChan <- neighbor
		}
	}

	return nil
}

func (n *NetNeighborScanner) GenerateNeighbor() {
	maxThread := 500
	threadchan := make(chan struct{}, maxThread)
	wait := sync.WaitGroup{}
	for _, netnode := range n.NetNodes {
		threadchan <- struct{}{}
		wait.Add(1)
		go func(node *NetNode) {
			err := n.scanNeighbor(node)
			if err != nil {
				fmt.Println(err)
			}
			wait.Done()
			<-threadchan
		}(netnode)
	}
	wait.Wait()
	n.ScanFinished = true

}

func (n *NetNeighborScanner) ReadChannel() {

	// read NetChassisIdChan
	go func() {
		for {
			netchassis := <-n.NetChassisIdChan
			n.NetChassisId.Set(netchassis[0], netchassis[1])
		}
	}()

	//valid neighbors info
	go func() {
		var mutex sync.Mutex
		for {
			neighbor := <-n.UnValidNeighborChan
			if rem_ip, ok := n.NetChassisId.Get(neighbor.RemoteIP); ok {
				fmt.Println("OK", rem_ip, n.NetChassisId.Data)
				neighbor.RemoteIP = rem_ip
				n.ValidNeighborChan <- neighbor
			} else {
				mutex.Lock()
				NodeListPush(n.UnValidNeighbor, neighbor)
				mutex.Unlock()
			}
		}
	}()

	go func() {
		var mutex sync.Mutex
		var neighbor *NetNeighbor
		epoch := 0
		removed := false // 当删除链表中的一个节点之后，更新此值
		for {
			current := n.UnValidNeighbor.Next
			for {
				if current.El == nil {
					if !removed {
						epoch += 1
						time.Sleep(1 * time.Second) //每次迭代延迟1s的时间
					}
					removed = false
					break
				}
				neighbor = current.El
				if rem_ip, ok := n.NetChassisId.Get(neighbor.RemoteIP); ok {
					neighbor.RemoteIP = rem_ip

					mutex.Lock()
					current.Prev.Next = current.Next
					current.Next.Prev = current.Prev
					mutex.Unlock()

					removed = true
					n.ValidNeighborChan <- neighbor
				}
				current = current.Next
			}

			if n.ScanFinished && epoch > 4 {
				break
			}
		}
	}()

}

func (n *NetNeighborScanner) SaveNeighbor(savefunc func(neighbor *NetNeighbor) error) {
	maxThread := 500
	threadchan := make(chan struct{}, maxThread)

	wait := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wait.Add(1)
		go func() {
			for {
				threadchan <- struct{}{}
				if len(n.ValidNeighborChan) == 0 {
					<-threadchan
					if n.ScanFinished {
						break
					}
					time.Sleep(2 * time.Second) //每次迭代延迟1s的时间
					continue
				}
				neighbor := <-n.ValidNeighborChan

				//执行回调函数
				if err := savefunc(neighbor); err != nil {
					fmt.Println(err)
				}

				<-threadchan
			}
			wait.Done()
		}()
	}
	wait.Wait()
	n.SaveFinished.Done()
}
