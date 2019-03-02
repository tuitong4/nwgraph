package scanner

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type NetNeighbor struct {
	localIP    string
	localPort  string
	remoteIP   string
	remotePort string
}

type NetNeighborScanner struct {
	NetnodeChan         chan *NetNode
	NetChassisIdChan    chan [2]string
	NetChassisId        SafeMap
	UnValidNeighborChan chan *NetNeighbor
	UnValidNeighbor     *NodeList
	ValidNeighborChan   chan *NetNeighbor
	Community           string
	GetNodeFinished     bool
	ScanFinished        bool
	SaveFinished        sync.WaitGroup
}

func (n *NetNeighborScanner) GetNetNodeMock(url string) error {
	/*
	* url is the NetNode infomaton data base on remote.
	 */
	for i := 1; i < 3; i++ {
		n.NetnodeChan <- &NetNode{
			level:      float64(i),
			mgt:        "172.0.0." + strconv.Itoa(i),
			oobmgt:     "",
			datacenter: "UNKOWN",
			vendor:     "H3C",
			model:      "12345",
			role:       "T1",
			service:    "UK",
			pod:        "POD001",
			name:       "SW",
			labels:     NodeLable["SWITCH"],
		}
	}
	n.GetNodeFinished = true
	return nil
}

func (n *NetNeighborScanner) GetNetNode(url string) error {
	/*
	* url is the NetNode infomaton data base on remote.
	 */

	resp, err := http.Get(url)
	if err != nil {
		n.GetNodeFinished = true
		return err
	}

	defer resp.Body.Close()
	var r = &RespBody{}
	if err := json.NewDecoder(resp.Body).Decode(r); err != nil {
		n.GetNodeFinished = true
		return err
	}
	if r.Code != 2000 {
		n.GetNodeFinished = true
		return fmt.Errorf("Failed retrive data for Code is %d", r.Code)
	}

	if r.Message != "OK" {
		n.GetNodeFinished = true
		return fmt.Errorf("Failed retrive data for Message is %s", r.Message)
	}
	for _, node := range r.Data.List {
		level := NodeLevel[node.Role]
		if node.Service == "BMC" {
			level += 1
		}
		if node.Service == "BSW" {
			level -= 0.3
		}

		var mgt string
		if node.ManagementIp == "" {
			if node.OutofbandIp != "" {
				mgt = node.OutofbandIp
			} else {
				continue
			}
		} else {
			mgt = node.ManagementIp
		}
		fmt.Println(mgt)
		n.NetnodeChan <- &NetNode{
			level:      level,
			mgt:        mgt,
			oobmgt:     node.OutofbandIp,
			datacenter: node.DatacenterShortName,
			vendor:     node.Manufacturer,
			model:      node.Model,
			role:       node.Role,
			service:    node.Service,
			pod:        node.PodName,
			name:       node.Name,
			labels:     NodeLable[node.Role],
		}
	}
	n.GetNodeFinished = true
	return nil
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
		n.NetChassisIdChan <- [2]string{id, nodehandler.node.mgt}
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

	for rem_idx, chassis := range rem_chassis {
		if localportname, ok := local_port[rem_idx]; ok {
			if rem_ip, ok := n.NetChassisId.Get(chassis); ok {
				pair := &NetNeighbor{
					localIP:    nodehandler.node.mgt,
					localPort:  localportname,
					remoteIP:   rem_ip,
					remotePort: rem_port[rem_idx],
				}
				n.ValidNeighborChan <- pair
			} else {
				pair := &NetNeighbor{
					localIP:    nodehandler.node.mgt,
					localPort:  localportname,
					remoteIP:   chassis,
					remotePort: rem_port[rem_idx],
				}
				n.UnValidNeighborChan <- pair
			}
		}
	}
	return nil
}

func (n *NetNeighborScanner) GenerateNeighbor() {
	maxThread := 500
	threadchan := make(chan struct{}, maxThread)
	wait := sync.WaitGroup{}
	for {
		if n.GetNodeFinished && len(n.NetnodeChan) == 0 {
			wait.Wait()
			n.ScanFinished = true
			break
		}
		threadchan <- struct{}{}
		netnode := <-n.NetnodeChan
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
			if rem_ip, ok := n.NetChassisId.Get(neighbor.remoteIP); ok {
				fmt.Println("OK", rem_ip, n.NetChassisId.Data)
				neighbor.remoteIP = rem_ip
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
				if rem_ip, ok := n.NetChassisId.Get(neighbor.remoteIP); ok {
					neighbor.remoteIP = rem_ip

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
