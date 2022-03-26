package scheduler

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/stretchr/testify/assert"
)

func TestZK_IAmLeader(t *testing.T) {
	timeoutProblem := "this test can fail if connect timeout is too slow for status/ state" +
		" to be updated. should be 5+ sec. if reset can't solve, you gotta do more debug :D"
	if testing.Short() {
		t.Skip("skipping TestZK_IAmLeader in short mode")
	}
	zkport := 2181
	conf := ZKLeaderElectorConfig{
		zkHosts:        []string{"127.0.0.1:" + strconv.Itoa(zkport)},
		electionName:   "",
		connectTimeout: 4 * time.Second,
	}
	electionName := "lala"
	conn1, _, err := zk.Connect(conf.zkHosts, conf.connectTimeout)
	if err != nil {
		log.Printf("Error in zk.Connect (%s): %v", conf.zkHosts, err)
	}
	defer conn1.Close()
	elector1 := NewZK(conn1, conf.Name("1"))
	conn2, _, err := zk.Connect(conf.zkHosts, conf.connectTimeout)
	if err != nil {
		log.Printf("Error in zk.Connect (%s): %v", conf.zkHosts, err)
	}
	defer conn2.Close()
	elector2 := NewZK(conn2, conf.Name("2"))
	{
		t.Log("one leader at a time")
		assert.Equal(t, true, elector1.IAmLeader(electionName, 0))
		assert.Equal(t, false, elector2.IAmLeader(electionName, 0))
	}
	{
		t.Log("node disconnect with zookeeper")
		elector1.conn.Close()
		time.Sleep(conf.connectTimeout) //only after this timeout is new leader elected
		assert.Equal(t, false, elector1.IAmLeader(electionName, 0), timeoutProblem)
		assert.Equal(t, true, elector2.IAmLeader(electionName, 0), timeoutProblem)
	}
	{
		t.Log("zookeeper is down")
		_ = exec.Command("bash", "-c", fmt.Sprintf("sudo iptables -I INPUT -p tcp --dport %v -j DROP", zkport)).Run()
		time.Sleep(conf.connectTimeout) //this depends on zk timeout to reset state. I think this is enough after multiple tries :D
		assert.Equal(t, false, elector1.IAmLeader(electionName, 0), timeoutProblem)
		assert.Equal(t, false, elector2.IAmLeader(electionName, 0), timeoutProblem)
		exec.Command("bash", "-c", fmt.Sprintf("sudo iptables -D INPUT -p tcp --dport %v -j DROP", zkport)).Run()
	}
	{
		t.Log("zookeeper is up again")
		time.Sleep(conf.connectTimeout) //this depends on zk to reconnect. I estimated it :D
		assert.Equal(t, true, elector2.IAmLeader(electionName, 0), timeoutProblem)
	}
}
