package main

/* In this file:
0. Globals: conns, logger, db, no, port
	Definitions: Record, RPCExt, RPCInt
1. Main (connet to db, init clocks, init keys entry, start up RPC server)
2. ConnectReplica (make conns to other replicas)
*/

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"sync"

	"../../util"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/savreline/GoVector/govec"
	"github.com/savreline/GoVector/govec/vclock"
)

// Global variables
var no int
var delay int
var noStr string
var port string
var pid string
var eLog string
var iLog string
var conns []*rpc.Client
var logger *govec.GoLog
var db *mongo.Database
var chans = make(map[chan vclock.VClock]chan vclock.VClock)
var lock = &sync.Mutex{}
var verbose = true

// Record is a DB Record
type Record struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

// RPCExt is the RPC object that receives commands from the driver
type RPCExt int

// RPCInt is the RPC Object for internal replica-to-replica communication
type RPCInt int

// Makes connection to the database, starts up the RPC server
func main() {
	/* Parse args, initialize data structures */
	noReplicas, _ := strconv.Atoi(os.Args[1])
	no, _ = strconv.Atoi(os.Args[2])
	noStr = os.Args[2]
	port = os.Args[3]
	dbPort := os.Args[4]
	delay, _ = strconv.Atoi(os.Args[5])
	conns = make([]*rpc.Client, noReplicas)

	/* Connect to MongoDB */
	dbClient, _ := util.Connect(dbPort)
	db = dbClient.Database("chev")
	util.PrintMsg(no, "Connected to DB on "+dbPort)

	/* Init vector clocks */
	pid = "R" + noStr
	logger = govec.InitGoVector(pid, pid, govec.GetDefaultConfig())

	/* Pre-allocate Keys entry */
	newRecord := Record{"Keys", []string{}}
	_, err := db.Collection("kvs").InsertOne(context.TODO(), newRecord)
	if err != nil {
		util.PrintErr(err)
	}

	/* Init RPC */
	server := rpc.NewServer()
	rpcext := new(RPCExt)
	rpcint := new(RPCInt)
	server.Register(rpcext)
	server.Register(rpcint)
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("listen error:", err)
	}

	/* Start Server */
	util.PrintMsg(no, "Server Listening on "+port)
	go server.Accept(l)
	select {}
}

// ConnectReplica connects this replica to others
func (t *RPCExt) ConnectReplica(args *util.ConnectArgs, reply *int) error {
	/* Parse Group Members */
	ports, _, err := util.ParseGroupMembersCVS("../driver/ports.csv", port)
	if err != nil {
		util.PrintErr(err)
	}

	/* Make RPC Connections */
	for i, port := range ports {
		conns[i] = util.RPCClient(port, "REPLICA "+strconv.Itoa(no)+": ")
	}

	return nil
}

// TerminateReplica writes to the log
func (t *RPCExt) TerminateReplica(args *util.ConnectArgs, reply *int) error {
	eLog = eLog + fmt.Sprint("Clock ", logger.GetCurrentVC())
	if verbose == true {
		err := ioutil.WriteFile("eRepl"+strconv.Itoa(no)+".txt", []byte(eLog), 0644)
		if err != nil {
			panic(err)
		}
		err = ioutil.WriteFile("iRepl"+strconv.Itoa(no)+".txt", []byte(iLog), 0644)
		if err != nil {
			panic(err)
		}
	}
	return nil
}

func broadcastClockValue(clockValue vclock.VClock) {
	for _, channel := range chans {
		channel <- clockValue
	}
}
