package mock

import (
	. "util"
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
		LocalPort:       map[string]string{"01": "F1/0/1", "02": "F1/0/2", "03": "F1/0/3"},
	},

	"172.0.0.2": Info{
		Mgt:             "172.0.0.2",
		ChassisID:       "aaaa-aaaa-aaaa-aa02",
		RemoteChassisID: map[string]string{"01": "aaaa-aaaa-aaaa-aa01", "02": "aaaa-aaaa-aaaa-aa01", "03": "aaaa-aaaa-aaaa-aa03"},
		LocalPort:       map[string]string{"01": "40GE1/0/1", "02": "40GE1/0/2", "03": "40GE1/0/3"},
		RemotePort:      map[string]string{"01": "F1/0/1", "02": "F1/0/2", "03": "F1/0/3"},
	},
	"172.0.0.3": Info{
		Mgt:             "172.0.0.3",
		ChassisID:       "aaaa-aaaa-aaaa-aa03",
		RemoteChassisID: map[string]string{"01": "aaaa-aaaa-aaaa-aa01", "02": "aaaa-aaaa-aaaa-aa01", "03": "aaaa-aaaa-aaaa-aa03"},
		LocalPort:       map[string]string{"01": "40GE1/0/1", "02": "40GE1/0/2", "03": "40GE1/0/3"},
		RemotePort:      map[string]string{"01": "F1/0/1", "02": "F1/0/2", "03": "F1/0/3"},
	},
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
	result = append(result, nwsw[n.node.Mgt].ChassisID)
	return result, nil
}
func (n *NetNodeHandler) RemChassisID() (map[string]string, error) {
	result := make(map[string]string)

	result = nwsw[n.node.Mgt].RemoteChassisID

	return result, nil
}

func (n *NetNodeHandler) RemPort() (map[string]string, error) {
	result := make(map[string]string)
	result = nwsw[n.node.Mgt].RemotePort
	return result, nil
}

func (n *NetNodeHandler) LocalPort() (map[string]string, error) {
	result := make(map[string]string)
	result = nwsw[n.node.Mgt].LocalPort
	return result, nil
}
