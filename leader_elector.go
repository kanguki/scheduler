package scheduler

import (
	"strconv"
	"strings"
	"time"

	"github.com/Comcast/go-leaderelection"
	le "github.com/Comcast/go-leaderelection"
	"github.com/go-zookeeper/zk"
)

type leaderElector interface {
	IAmLeader(job string) bool
}

//zookeeper-based
const (
	zkElectionRoot     = "/scheduler_elections"
	zkElectionRootData = "elections for scheduled jobs"
)

type ZKLeaderElector struct {
	conn *zk.Conn
	core *le.Election
	ZKLeaderElectorConfig
	status chan le.Status
}
type ZKLeaderElectorConfig struct {
	id             string
	zkHosts        []string
	electionName   string
	connectTimeout time.Duration //estimate time zk need, should be set from 5+ secs
}

func (c ZKLeaderElectorConfig) Name(id string) ZKLeaderElectorConfig {
	c.id = id
	return c
}
func NewZK(zkConn *zk.Conn, conf ZKLeaderElectorConfig) *ZKLeaderElector {
	return &ZKLeaderElector{
		conn:                  zkConn,
		core:                  &le.Election{},
		ZKLeaderElectorConfig: conf,
		status:                make(chan le.Status, 1),
	}
}

//version is unused for now
func (zl *ZKLeaderElector) IAmLeader(electionName string, version int) bool {
	handleNoElectionIsRunning := func() bool {
		Log("%v handleNoElectionIsRunning", zl.id)
		zl.removeCurrentStatus()
		go zl.elect(electionName, version)
		select {
		case status := <-zl.status:
			zl.status <- status
			if status.Role == le.Leader {
				return true
			}
			return false
		case <-time.After(zl.connectTimeout):
			Log("Timeout Decide Leader for %v!!!", zl.id)
			return false
		}
	}
	select {
	case status := <-zl.status:
		zl.status <- status
		// Log("%v connection state %v", zl.id, zl.conn.State().String())
		if zl.conn.State() != zk.StateHasSession { //if no check zk connection, it might be caching old status, and there might be multiple leaders at a time
			return handleNoElectionIsRunning()
		}
		if status.Role == le.Leader {
			return true
		}
		if status.NowFollowing == "" { //I'm not leader, no one is leader now
			return handleNoElectionIsRunning()
		}
		return false

	default:
		return handleNoElectionIsRunning()
	}

}

func (zl *ZKLeaderElector) elect(electionName string, version int) {
	zl.removeCurrentStatus()
	var err error
	path := zkElectionRoot + "/" + electionName
	// Create a persistent znode as election node in ZooKeeper
	{
		_, err := zl.conn.Create(zkElectionRoot, []byte(zkElectionRootData), 0, zk.WorldACL(zk.PermAll))
		if err != nil && !strings.Contains(err.Error(), "node already exists") {
			Log("%v Error creating the election node <%s>: %v", zl.id, zkElectionRoot, err)
			zl.status <- le.Status{} //return a zero object to make this deciding process faster
			return
		}
		_, err = zl.conn.Create(path, []byte(strconv.Itoa(version)), 0, zk.WorldACL(zk.PermAll))
		if err != nil && !strings.Contains(err.Error(), "node already exists") {
			Log(" %v Error creating the election node <%s>: %v", zl.id, path, err)
			zl.status <- le.Status{} //return a zero object to make this deciding process faster
			return
		}
	}

	//start election and watch event
	{
		// Log("%v connection state %v", zl.id, zl.conn.State().String())
		zl.core, err = le.NewElection(zl.conn, path, zl.id)
		if err != nil {
			Log("%v cannot start an election %v", zl.id, err)
			zl.status <- le.Status{} //return a zero object to make this deciding process faster
			return
		}
		go zl.core.ElectLeader()
		for {
			select {
			case status, ok := <-zl.core.Status():
				zl.removeCurrentStatus()
				Log("Candidate <%v> received status message: <%v>.", zl.id, status)
				if !ok {
					Log("%v channel closed, election is terminated!!!", zl.id)
					zl.core.Resign()
					return
				}
				if status.Err != nil {
					Log("%v received election status error <<%v>>", zl.id, status.Err)
					zl.core.Resign()
					return
				}
				zl.status <- status
				if status.Role == leaderelection.Leader {
					Log("%v is leader", zl.id)
				}
			}
		}
	}

}
func (zl *ZKLeaderElector) removeCurrentStatus() {
	select {
	case <-zl.status: //just pop out whatever in current status
	default:
	}
}
