package util

type NetNode struct {
	Id         int64
	Level      float64
	Mgt        string
	Oobmgt     string
	Datacenter string
	Vendor     string
	Model      string
	Role       string
	Service    string
	Pod        string
	Name       string
	Labels     []string
}

//对于没法将nodeid转换为int64的，在GLOBAL_CONUTER中选择一个数
var GLOBAL_CONUTER = int64(1000)

//记录从GLOBAL_CONUTER选择的节点信息
var GLOBAL_NODEIDS = map[string]int64{}

func GenNodeID(id string) int64 {
	ipv4, err := IPv4(id)
	if err != nil {
		if _, ok := GLOBAL_NODEIDS[id]; !ok {
			GLOBAL_CONUTER += 1
			GLOBAL_NODEIDS[id] = GLOBAL_CONUTER
			return GLOBAL_CONUTER
		}
		return GLOBAL_NODEIDS[id]
	}
	return int64(ipv4.Numeric())
}
