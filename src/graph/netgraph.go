package graph

import (
	"fmt"
	"neo4j-go-driver/neo4j"
	"strings"
	. "util"
)

type NetGraph struct {
	session      neo4j.Session
	driver       neo4j.Driver
	neo4jserver  string
	neo4jauth    neo4j.AuthToken
	accessmode   neo4j.AccessMode // 0 is WriteMode, 1 is ReadMode
	tx           neo4j.Transaction
	nodeidxcache map[int64]int64
}

//NewNetGraph
func NewNetGraph(bolt, boltname, boltpasswd string, accessmode neo4j.AccessMode) *NetGraph {
	return &NetGraph{
		session:      nil,
		driver:       nil,
		neo4jserver:  bolt,
		neo4jauth:    neo4j.BasicAuth(boltname, boltpasswd, ""),
		accessmode:   accessmode,
		tx:           nil,
		nodeidxcache: map[int64]int64{},
	}
}

func (n *NetGraph) ConnectNeo4j() error {
	var err error
	if n.driver, err = neo4j.NewDriver(n.neo4jserver, n.neo4jauth); err != nil {
		return err
	}

	if n.session, err = n.driver.Session(n.accessmode); err != nil {
		return err
	}

	return nil
}

func (n *NetGraph) TxStart() error {
	tx, err := n.session.BeginTransaction()
	if err != nil {
		n.tx = nil
		return err
	}
	n.tx = tx
	return nil
}
func (n *NetGraph) TxClose() error {
	if n.tx == nil {
		return nil
	}
	err := n.tx.Close()
	if err != nil {
		return err
	}
	n.tx = nil
	return nil
}

func (n *NetGraph) TxCommit() error {
	if n.tx == nil {
		return nil
	}
	err := n.tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (n *NetGraph) Exit() {
	_ = n.session.Close()
	_ = n.driver.Close()
}

func (n *NetGraph) CreateNetNode(node *NetNode) error {
	params := map[string]interface{}{
		"id":      node.Id,
		"level":   node.Level,
		"mgt":     node.Mgt,
		"oobmgt":  node.Oobmgt,
		"dc":      node.Datacenter,
		"vendor":  node.Vendor,
		"model":   node.Model,
		"role":    node.Role,
		"service": node.Service,
		"pod":     node.Pod,
		"name":    node.Name,
	}

	statement := `CREATE(n:` + strings.Join(node.Labels, ":") +
		`{id:$id ,level:$level, mgt:$mgt, oobmgt:$oobmgt, dc:$dc,` +
		`vendor:$vendor, model:$model, role:$role, service:$service,` +
		`pod:$pod, name:$name}) RETURN id(n)`

	result, err := n.session.Run(statement, params)

	if err != nil {
		return err
	}

	for result.Next() {
		r, ok := result.Record().GetByIndex(0).(int64)

		if !ok {
			return fmt.Errorf("unconvert record type to int64")
		}
		n.nodeidxcache[node.Id] = r
	}

	return nil
}

func (n *NetGraph) CreateNetNodeWithTx(node *NetNode) error {
	params := map[string]interface{}{
		"id":      node.Id,
		"level":   node.Level,
		"mgt":     node.Mgt,
		"oobmgt":  node.Oobmgt,
		"dc":      node.Datacenter,
		"vendor":  node.Vendor,
		"model":   node.Model,
		"role":    node.Role,
		"service": node.Service,
		"pod":     node.Pod,
		"name":    node.Name,
	}

	statement := `CREATE(n:` + strings.Join(node.Labels, ":") +
		`{id:$id ,level:$level, mgt:$mgt, oobmgt:$oobmgt, dc:$dc,` +
		`vendor:$vendor, model:$model, role:$role, service:$service,` +
		`pod:$pod, name:$name}) RETURN id(n)`

	result, err := n.tx.Run(statement, params)

	if err != nil {
		return err
	}

	for result.Next() {
		r, ok := result.Record().GetByIndex(0).(int64)

		if !ok {
			return fmt.Errorf("unconvert record type to int64")
		}
		n.nodeidxcache[node.Id] = r
	}

	return nil
}

func (n *NetGraph) CreateNetLinkByID(startid, endid int64, localports, remoteports []string) error {
	params := map[string]interface{}{
		"start":  startid,
		"end":    endid,
		"lports": localports,
		"rports": remoteports,
	}

	_, err := n.session.Run(
		`MATCH(s), (e) WHERE id(s)=$start and id(e)=$end CREATE(s)-[:LINK_TO{lports:$lports, rports:$rports}]->(e)`, params)

	if err != nil {
		return err
	}

	return nil
}

func (n *NetGraph) CreateNetLinkByIDWithTX(startid, endid int64, localports, remoteports []string) error {
	params := map[string]interface{}{
		"start":  startid,
		"end":    endid,
		"lports": localports,
		"rports": remoteports,
	}

	_, err := n.tx.Run(
		`MATCH(s), (e) WHERE id(s)=$start and id(e)=$end CREATE(s)-[:LINK_TO{lports:$lports, rports:$rports}]->(e)`, params)

	if err != nil {
		return err
	}

	return nil
}

func (n *NetGraph) CreateNetLinkByNetNodeID(startid, endid int64, localports, remoteports []string) error {
	params := map[string]interface{}{
		"start":  startid,
		"end":    endid,
		"lports": localports,
		"rports": remoteports,
	}

	_, err := n.session.Run(
		`MATCH(s:SWITCH{id:$start}), (e:SWITCH{id:$end}) CREATE(s)-[:LINK_TO{lports:$lports, rports:$rports}]->(e)`, params)

	if err != nil {
		return err
	}

	return nil
}

func (n *NetGraph) CreateNetLinkByNetNodeIDWithTX(startid, endid int64, localports, remoteports []string) error {
	params := map[string]interface{}{
		"start":  startid,
		"end":    endid,
		"lports": localports,
		"rports": remoteports,
	}

	_, err := n.tx.Run(
		`MATCH(s:SWITCH{id:$start}), (e:SWITCH{id:$end}) CREATE(s)-[:LINK_TO{lports:$lports, rports:$rports}]->(e)`, params)

	if err != nil {
		return err
	}

	return nil
}

func (n *NetGraph) QueryNetNode(props map[string]interface{}) ([]neo4j.Node, error) {
	/*
	* Make sure the key of map is same as the NetNode's property name.
	 */
	filter := make([]string, 0, 1)
	for k, _ := range props {
		filter = append(filter, k+":$"+k)
	}
	statement := "MATCH(n:SWITCH{" + strings.Join(filter, ",") + "}) RETURN (n)"

	result, err := n.session.Run(statement, props)

	if err != nil {
		return nil, err
	}

	nodes := []neo4j.Node{}
	for result.Next() {
		r := result.Record().Values()
		if len(r) == 0 {
			nodes = append(nodes, nil)
		} else if len(r) == 1 {
			node, err := neo4j.ParaseNodeValue(r[0])
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, node)
		} else {
			return nil, fmt.Errorf("Unformated result")
		}
	}

	return nodes, nil
}

func (n *NetGraph) QueryNetLink(startlable, endlable []string, start, end map[string]interface{}, direction string) ([]neo4j.Relationship, error) {
	/*
	* start and end is the filter props of NetNodes.
	* Make sure the key of start and end is same as the NetNode's property name.
	* direction is the relationship direction.
	* direction value is "--", "->", "<-", which indicates the relationship type of tow nodes
	 */
	if len(startlable) == 0 || startlable == nil {
		startlable = []string{"SWITCH"}
	}

	if len(endlable) == 0 || startlable == nil {
		endlable = []string{"SWITCH"}
	}

	params := map[string]interface{}{}
	startfilter := make([]string, 0, 1)
	for k, v := range start {
		startfilter = append(startfilter, k+":$s"+k)
		params["s"+k] = v
	}
	statement_start := "MATCH(n:" + strings.Join(startlable, ":") + "{" + strings.Join(startfilter, ",") + "})"

	endfilter := make([]string, 0, 1)
	for k, v := range end {
		endfilter = append(endfilter, k+":$e"+k)
		params["e"+k] = v
	}

	statement_end := "(e:" + strings.Join(startlable, ":") + "{" + strings.Join(endfilter, ",") + "})"

	var statement string
	if direction == "->" {
		statement = statement_start + "-[r:LINK_TO]->" + statement_end + "RETURN r"
	} else if direction == "--" {
		statement = statement_start + "-[r:LINK_TO]-" + statement_end + "RETURN r"
	} else if direction == "<-" {
		statement = statement_start + "<-[r:LINK_TO]-" + statement_end + "RETURN r"
	} else {
		return nil, fmt.Errorf("Unspport direction '%s'", direction)
	}
	result, err := n.session.Run(statement, params)
	if err != nil {
		return nil, err
	}

	links := []neo4j.Relationship{}
	for result.Next() {
		r := result.Record().Values()
		if len(r) == 0 {
			links = append(links, nil)
		} else if len(r) == 1 {
			link, err := neo4j.ParseRelationshipValue(r[0])
			if err != nil {
				return nil, err
			}
			links = append(links, link)
		} else {
			return nil, fmt.Errorf("Unformated result")
		}
	}

	return links, nil
}
