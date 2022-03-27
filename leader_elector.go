package scheduler

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Comcast/go-leaderelection"
	le "github.com/Comcast/go-leaderelection"
	"github.com/go-zookeeper/zk"
	"github.com/google/uuid"
)

type LeaderElector interface {
	IAmLeader(electionName string, version int) bool
	//close network calls, open files,... defer in main goroutine
	CleanResource()
}

const (
	zkElectionRoot     = "/scheduler_elections"
	zkElectionRootData = "elections for scheduled jobs"
)

func NewLeaderElector() LeaderElector {
	elector := os.Getenv(SCHEDULER_ELECTOR)
	if elector == "" {
		return &SingleNodeElector{}
	} else {
		switch elector {
		case "zk":
			return NewZK(nil, ZKLeaderElectorConfig{})
		default:
			Log("missing variables %v and its type to init an elector in multi-node state", SCHEDULER_ELECTOR)
			return nil
		}
	}
}

//mock, always return true, as there's no need to elect leader in a single node case
type SingleNodeElector struct{}

func (zl *SingleNodeElector) IAmLeader(electionName string, version int) bool {
	return true
}
func (zl *SingleNodeElector) CleanResource() {}

//zookeeper-based
type ZKLeaderElector struct {
	conn *zk.Conn
	core *le.Election
	ZKLeaderElectorConfig
	status chan le.Status
}
type ZKLeaderElectorConfig struct {
	Id             string
	zkHosts        []string
	connectTimeout time.Duration //estimate time zk need, should be set from 5+ secs
}

func NewZK(zkConn *zk.Conn, conf ZKLeaderElectorConfig) *ZKLeaderElector {
	var err error
	if conf.zkHosts == nil { //empty config
		zkUrls := os.Getenv(ZOOKEEPER_URLS)
		if zkUrls == "" {
			Log("Empty %v, can't connect to zookeeper", ZOOKEEPER_URLS)
			return nil
		}
		conf = ZKLeaderElectorConfig{
			Id:             uuid.New().String(),
			zkHosts:        strings.Split(zkUrls, ","),
			connectTimeout: 3 * time.Second,
		}
	}
	if zkConn == nil {
		zkConn, _, err = zk.Connect(conf.zkHosts, conf.connectTimeout)
		if err != nil {
			Log("Error in zk.Connect (%s): %v", conf.zkHosts, err)
			return nil
		}
	}
	Log("Create ZK elector %v", conf.Id)
	return &ZKLeaderElector{
		conn:                  zkConn,
		core:                  &le.Election{},
		ZKLeaderElectorConfig: conf,
		status:                make(chan le.Status, 1),
	}
}

func (zl *ZKLeaderElector) CleanResource() {
	Log("ZKLeaderElector is cleaning resources...")
	zl.conn.Close()
	zl.cleanStatusAndResign()
}

//param version is unused for now
func (zl *ZKLeaderElector) IAmLeader(electionName string, version int) bool {
	handleNoElectionIsRunning := func() bool {
		zl.cleanStatusAndResign()
		go zl.elect(electionName, version)
		select {
		case status := <-zl.status:
			zl.status <- status
			if status.Role == le.Leader {
				return true
			}
			return false
		case <-time.After(zl.connectTimeout):
			Log("Timeout Decide Leader for %v! There may be a network problem", zl.Id)
			zl.cleanStatusAndResign()
			return false
		}
	}
	select {
	case status := <-zl.status:
		zl.status <- status
		if zl.conn.State() != zk.StateHasSession { //if there's a network partition, it might be caching old status, and there might be multiple leaders at a time
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
	var err error
	path := zkElectionRoot + "/" + electionName
	// Create a persistent znode as election node in ZooKeeper
	{
		errPath := zkElectionRoot
		_, err = zl.conn.Create(zkElectionRoot, []byte(zkElectionRootData), 0, zk.WorldACL(zk.PermAll))
		_, err2 := zl.conn.Create(path, []byte(strconv.Itoa(version)), 0, zk.WorldACL(zk.PermAll))
		if err2 != nil {
			err = err2
			errPath = path
		}
		if err != nil && !strings.Contains(err.Error(), "node already exists") {
			Log("%v Error creating the election node %s: %v", zl.Id, errPath, err)
			zl.cleanStatusAndResign()
			zl.status <- le.Status{} //return a zero object so this doesn't have to wait till timeout to proceed
			return
		}
	}

	//start election and watch event
	{
		zl.core, err = le.NewElection(zl.conn, path, zl.Id)
		if err != nil {
			Log("%v cannot start an election %v", zl.Id, err)
			zl.cleanStatusAndResign()
			zl.status <- le.Status{}
			return
		}
		go zl.core.ElectLeader()
		for {
			select {
			case status, ok := <-zl.core.Status():
				Log("Candidate %v received status message: %v.", zl.Id, status)
				if !ok {
					Log("%v channel closed, election is terminated!!!", zl.Id)
					zl.cleanStatusAndResign()
					zl.status <- le.Status{}
					return
				}
				if status.Err != nil {
					Log("%v received election status error %v", zl.Id, status.Err)
					zl.cleanStatusAndResign()
					zl.status <- le.Status{}
					return
				}
				{
					//new status, clear old cache
					select {
					case <-zl.status:
					default:
					}
					zl.status <- status
				}
				if status.Role == leaderelection.Leader {
					Log("%v is leader of %v", zl.Id, zl.core.ElectionResource)
				}
			}
		}
	}

}

//update status from library is slow some times, so i close connection directly
func (zl *ZKLeaderElector) cleanStatusAndResign() {
	select {
	case status := <-zl.status:
		zl.conn.Close()
		zkConn, _, err := zk.Connect(zl.zkHosts, zl.connectTimeout)
		if err != nil {
			Log("Error in zk.Connect (%s): %v", zl.zkHosts, err)
		}
		zl.conn = zkConn
		if status.CandidateID != "" { //old status cached.
			select {
			case s := <-zl.core.Status():
				if s != (le.Status{}) { //Status is not closed
					Log("%v resign", status.CandidateID)
					zl.core.Resign()
				}
			default:
			}
		}
	default:
	}
}
