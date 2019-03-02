package scanner

import (
	"encoding/hex"
	"github.com/gosnmp"
	"strings"
	"time"
)

const (
	lldpLocChassisID      = "1.0.8802.1.1.2.1.3.2.0"
	lldpLocChassisIDNexus = "1.3.6.1.2.1.2.2.1.6"
	lldpRemChassisID      = "1.0.8802.1.1.2.1.4.1.1.5"
	lldpRemPortID         = "1.0.8802.1.1.2.1.4.1.1.7"
	lldpLocPortID         = "1.0.8802.1.1.2.1.3.7.1.3"
)

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
	node  *NetNode
	snmpd *gosnmp.GoSNMP
}

func NewNetNodeHandler(netnode *NetNode, community string) *NetNodeHandler {
	return &NetNodeHandler{
		node: netnode,
		snmpd: &gosnmp.GoSNMP{
			Target:         netnode.mgt,
			Port:           uint16(161),
			Community:      community,
			Version:        gosnmp.Version2c,
			Retries:        1,
			Timeout:        time.Duration(3) * time.Second,
			MaxRepetitions: 3},
	}
}

func (n *NetNodeHandler) SNMPConnect() error {
	return n.snmpd.Connect()
}

func (n *NetNodeHandler) SNMPClose() error {
	return n.snmpd.Conn.Close()
}

func (n *NetNodeHandler) SelfChassisID() ([]string, error) {
	var result = []string{}
	if strings.Contains(n.node.model, "Nexus") {
		resp, err := n.snmpd.BulkWalkAll(lldpLocChassisIDNexus)
		if err != nil {
			return nil, err
		}
		for _, pdu := range resp {
			switch pdu.Type {
			case gosnmp.OctetString:
				result = append(result, hex.EncodeToString(pdu.Value.([]byte)))
			}
		}

	} else {
		_resp, err := n.snmpd.Get([]string{lldpLocChassisID})
		if err != nil {
			return nil, err
		}

		resp := _resp.Variables[0]

		switch resp.Type {
		case gosnmp.OctetString:
			result = append(result, hex.EncodeToString(resp.Value.([]byte)))
		}
	}

	return result, nil
}
func (n *NetNodeHandler) RemChassisID() (map[string]string, error) {
	result := make(map[string]string)
	resp, err := n.snmpd.BulkWalkAll(lldpRemChassisID)
	if err != nil {
		return nil, err
	}

	for _, pdu := range resp {
		parts := strings.Split(pdu.Name, ".")
		index := parts[len(parts)-2]

		switch pdu.Type {
		case gosnmp.OctetString:
			h := hex.EncodeToString(pdu.Value.([]byte))
			if len(h) == 12 {
				result[index] = h
			}
		}
	}
	return result, err
}

func (n *NetNodeHandler) RemPort() (map[string]string, error) {
	result := make(map[string]string)
	resp, err := n.snmpd.BulkWalkAll(lldpRemPortID)
	if err != nil {
		return nil, err
	}

	for _, pdu := range resp {
		parts := strings.Split(pdu.Name, ".")
		index := parts[len(parts)-2]

		switch pdu.Type {
		case gosnmp.OctetString:
			result[index] = string(pdu.Value.([]byte))

		}
	}
	return result, err
}

func (n *NetNodeHandler) LocalPort() (map[string]string, error) {
	result := make(map[string]string)

	resp, err := n.snmpd.BulkWalkAll(lldpLocPortID)
	if err != nil {
		return nil, err
	}

	for _, pdu := range resp {
		parts := strings.Split(pdu.Name, ".")
		index := parts[len(parts)-1]

		switch pdu.Type {
		case gosnmp.OctetString:
			result[index] = string(pdu.Value.([]byte))

		}
	}
	return result, err
}
