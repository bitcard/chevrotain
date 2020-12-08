package main

import (
	"io/ioutil"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"sync"

	"../util"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/savreline/GoVector/govec"
)

// Constants
const (
	dCollection = "kvsd"
	sCollection = "kvs"
)

// Global variables
var no int
var noStr string
var ports []string
var ips []string
var eLog string
var iLog string
var verbose = true      // print to info console?
var conns []*rpc.Client // RPC connections to other replicas
var db *mongo.Database
var logger *govec.GoLog
var id int    // unique ids associated with elements
var delay int // emulated link delay

// Map of channels that are used for communication with waiting RPC Calls
// and the associated lock
var chans map[chan bool]chan bool
var lock sync.Mutex

// RPCExt is the RPC object that receives commands from the client
type RPCExt int

// RPCInt is the RPC object for internal replica-to-replica communication
type RPCInt int

// Makes connection to the database, initializes data structures, starts up the RPC server
func main() {
	var err error

	/* Parse command link arguments */
	no, err = strconv.Atoi(os.Args[1])
	noStr = os.Args[1]
	port := os.Args[2]
	dbPort := os.Args[3]
	delay, err = strconv.Atoi(os.Args[4])
	if err != nil {
		util.PrintErr(noStr, err)
	}

	/* Parse group member information */
	ips, ports, _, err = util.ParseGroupMembersCVS("../ports.csv", port)
	if err != nil {
		util.PrintErr(noStr, err)
	}
	noReplicas := len(ports) + 1

	/* Init data structures */
	conns = make([]*rpc.Client, noReplicas)
	chans = make(map[chan bool]chan bool)
	id = no * 100000

	/* Init vector clocks */
	logger = govec.InitGoVector("R"+noStr, "R"+noStr, govec.GetDefaultConfig())

	/* Connect to MongoDB */
	dbClient, _ := util.ConnectDb(noStr, "localhost", dbPort)
	db = dbClient.Database("chev")
	util.PrintMsg(noStr, "Connected to DB on "+dbPort)

	/* Init RPC */
	rpcext := new(RPCExt)
	rpcint := new(RPCInt)
	rpc.Register(rpcint)
	rpc.Register(rpcext)
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		util.PrintErr(noStr, err)
	}

	/* Start server */
	util.PrintMsg(noStr, "RPC Server Listening on "+port)
	go rpc.Accept(l)
	select {}
}

// InitReplica connects this replica to others
func (t *RPCExt) InitReplica(args *util.InitArgs, reply *int) error {
	for i, port := range ports {
		conns[i] = util.RPCClient(noStr, ips[i], port)
	}
	return nil
}

// TerminateReplica generates the "lookup" view collection of the database
// and saves the logs to disk
func (t *RPCExt) TerminateReplica(args *util.RPCExtArgs, reply *int) error {
	lookup()
	if verbose {
		err := ioutil.WriteFile("Repl"+noStr+".txt", []byte(eLog), 0644)
		if err != nil {
			util.PrintErr(noStr, err)
		}
	}
	return nil
}
