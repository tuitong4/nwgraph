package mock

import (
	"strconv"
	"sync"
	. "util"
)

/*
* 一个自定义链表，用于频繁的操作删除
 */
type NodeList struct {
	El   *NetNeighbor
	Prev *NodeList
	Next *NodeList
}

func NodeListInit() (head *NodeList) {
	head = &NodeList{
		El:   nil,
		Prev: nil,
		Next: nil,
	}
	head.Prev = head
	head.Next = head
	return
}

func NodeListPush(head *NodeList, el *NetNeighbor) {
	newEl := &NodeList{
		El:   el,
		Prev: head.Prev,
		Next: head,
	}
	head.Prev.Next = newEl
	head.Prev = newEl
}

type SafeMap struct {
	Lock sync.RWMutex
	Data map[string]string
}

/*
* 一个自定义Map，用于并发读写
 */

func NewSafeMap(cap int) SafeMap {
	return SafeMap{
		Lock: sync.RWMutex{},
		Data: make(map[string]string, cap),
	}
}

func (m *SafeMap) Set(key string, val string) {
	m.Lock.Lock()
	m.Data[key] = val
	m.Lock.Unlock()
}

func (m *SafeMap) Get(key string) (val string, ok bool) {
	m.Lock.RLock()
	val, ok = m.Data[key]
	m.Lock.RUnlock()
	return val, ok
}

/*
* 用于解析API数据的结构体
 */

type RespBody struct {
	Code    int64
	Data    DataBlock
	Message string
}

type DataBlock struct {
	List       []ListBlock
	TotalCount float64
}

type ListBlock struct {
	ID                  string      `json:"id"`
	SID                 string      `json:"sid"`
	Name                string      `json:"name"`
	Describe            string      `json:"describe"`
	Type                string      `json:"type"`
	SN                  string      `json:"sn"`
	AssetId             string      `json:"asset_id"`
	Role                string      `json:"role"`
	StackRole           string      `json:"stack_role"`
	MemberId            int64       `json:"member_id"`
	ServiceStatus       string      `json:"service_status"`
	State               string      `json:"state"`
	RaState             string      `json:"ra_state"`
	MonitorState        string      `json:"monitor_state"`
	Constructed         int64       `json:"constructed"`
	Business            string      `json:"business"`
	Service             string      `json:"service"`
	ManagementIpId      string      `json:"management_ip_id"`
	ManagementIp        string      `json:"management_ip"`
	OutofbandIpId       string      `json:"outofband_ip_id"`
	OutofbandIp         string      `json:"outofband_ip"`
	SoftVersion         string      `json:"soft_version"`
	PatchVersion        string      `json:"patch_version"`
	Manufacturer        string      `json:"manufacturer"`
	Brand               string      `json:"brand"`
	Model               string      `json:"model"`
	EndofLife           string      `json:"end_of_life"`
	DatacenterId        string      `json:"datacenter_id"`
	DatacenterName      string      `json:"datacenter_name"`
	DatacenterShortName string      `json:"datacenter_short_name"`
	RoomID              string      `json:"room_id"`
	RoomName            string      `json:"room_name"`
	RackId              string      `json:"rack_id"`
	RackName            string      `json:"rack_name"`
	PodId               string      `json:"pod_id"`
	PodName             string      `json:"pod_name"`
	PodMode             string      `json:"pod_mode"`
	StackId             string      `json:"stack_id"`
	Extra               interface{} `json:"extra"`
	CreatedTime         string      `json:"created_time"`
	UpdatedTime         string      `json:"updated_time"`
}

/*
* Node的一些基本信息结构
 */
var NodeLevel = map[string]float64{
	"T0": 3,
	"T1": 2,
	"T2": 1,
	"WE": 0,
	"LE": 0,
	"DE": -1,
	"WR": -2,
	"LR": -2,
	"PR": -3,
	"GR": -4,
}

var NodeLable = map[string][]string{
	"T0": {"SWITCH"},
	"T1": {"SWITCH"},
	"T2": {"SWITCH"},
	"WE": {"SWITCH"},
	"LE": {"SWITCH", "DCI"},
	"WR": {"SWITCH", "BACKBONE", "DCI"},
	"LR": {"SWITCH", "BACKBONE", "DCI"},
	"PR": {"SWITCH", "BACKBONE"},
	"GR": {"SWITCH", "BACKBONE", "DCI"},
}

func GetNetNodeMock(url string) ([]*NetNode, error) {
	/*
	* url is the NetNode infomaton data base on remote.
	 */
	var nodes = make([]*NetNode, 0, 10)
	for i := 1; i < 4; i++ {
		nodes = append(nodes, &NetNode{
			Id:         GenNodeID("172.0.0." + strconv.Itoa(i)),
			Level:      float64(i),
			Mgt:        "172.0.0." + strconv.Itoa(i),
			Oobmgt:     "",
			Datacenter: "UNKOWN",
			Vendor:     "H3C",
			Model:      "12345",
			Role:       "T1",
			Service:    "UK",
			Pod:        "POD001",
			Name:       "SW",
			Labels:     NodeLable["SWITCH"],
		})
	}
	return nodes, nil
}
