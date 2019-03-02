package mock

import (
	"strconv"
)

const (
	lldpLocChassisID      = "1.0.8802.1.1.2.1.3.2.0"
	lldpLocChassisIDNexus = "1.3.6.1.2.1.2.2.1.6"
	lldpRemChassisID      = "1.0.8802.1.1.2.1.4.1.1.5"
	lldpRemPortID         = "1.0.8802.1.1.2.1.4.1.1.7"
	lldpLocPortID         = "1.0.8802.1.1.2.1.3.7.1.3"
)

type Info struct {
	Mgt             string
	ChassisID       string
	RemoteChassisID map[string]string
	RemotePort      map[string]string
	LocalPort       map[string]string
}

var nwsw = map[string]Info{
	"172.0.0.1": Info{
		Mgt:             "172.0.0.1",
		ChassisID:       "aaaa-aaaa-aaaa-aa01",
		RemoteChassisID: map[string]string{"01": "aaaa-aaaa-aaaa-aa01", "02": "aaaa-aaaa-aaaa-aa02", "03": "aaaa-aaaa-aaaa-aa03"},
		RemotePort:      map[string]string{"01": "40GE1/0/1", "02": "40GE1/0/1", "03": "40GE1/0/3"},
		LocalPort:       map[string]string{"01": "F1/0/1", "02": "F1/0/1", "03": "F1/0/3"},
	},

	"172.0.0.2": Info{
		Mgt:             "172.0.0.2",
		ChassisID:       "aaaa-aaaa-aaaa-aa02",
		RemoteChassisID: map[string]string{"01": "aaaa-aaaa-aaaa-aa01", "02": "aaaa-aaaa-aaaa-aa01", "03": "aaaa-aaaa-aaaa-aa03"},
		LocalPort:       map[string]string{"01": "40GE1/0/1", "02": "40GE1/0/1", "03": "40GE1/0/3"},
		RemotePort:      map[string]string{"01": "F1/0/1", "02": "F1/0/1", "03": "F1/0/3"},
	},
}

type NetNode struct {
	id         int64
	level      float64
	mgt        string
	oobmgt     string
	datacenter string
	vendor     string
	model      string
	role       string
	service    string
	pod        string
	name       string
	labels     []string
}

type NetNodeHandler struct {
	node *NetNode
}

func NewNetNodeHandler(netnode *NetNode, community string) *NetNodeHandler {
	return &NetNodeHandler{
		node: netnode,
	}
}

func (n *NetNodeHandler) SNMPConnect() error {
	return nil
}

func (n *NetNodeHandler) SNMPClose() error {
	return nil
}

func (n *NetNodeHandler) SelfChassisID() ([]string, error) {
	var result = []string{}
	result = append(result, nwsw[n.node.mgt].ChassisID)
	return result, nil
}
func (n *NetNodeHandler) RemChassisID() (map[string]string, error) {
	result := make(map[string]string)

	result = nwsw[n.node.mgt].RemoteChassisID

	return result, nil
}

func (n *NetNodeHandler) RemPort() (map[string]string, error) {
	result := make(map[string]string)
	result = nwsw[n.node.mgt].RemotePort
	return result, nil
}

func (n *NetNodeHandler) LocalPort() (map[string]string, error) {
	result := make(map[string]string)
	result = nwsw[n.node.mgt].LocalPort
	return result, nil
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
