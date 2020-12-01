package util

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/rpc"
	"os"
	"time"

	"github.com/savreline/GoVector/govec/vclock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CmRecord is a CmRDT DB Record
type CmRecord struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

// CvRecord is a CvRDT DB Record
type CvRecord struct {
	Name   string       `json:"name"`
	Values []ValueEntry `json:"values"`
}

// ValueEntry is a value along with the timestamp
type ValueEntry struct {
	Value     string        `json:"name"`
	Timestamp vclock.VClock `json:"time"`
}

// RPCExtArgs are the arguments to any RPCExt Call
type RPCExtArgs struct {
	Key, Value string
}

// InitArgs are the arguments to Init RPCExt Call
type InitArgs struct {
	Settings [2]int
	TimeInt  int
}

// ConnectDb to MongoDB on the given port, as per https://www.mongodb.com/golang
func ConnectDb(no string, port string) (*mongo.Client, context.Context) {
	urlString := "mongodb://localhost:" + port + "/"

	client, err := mongo.NewClient(options.Client().ApplyURI(urlString))
	if err != nil {
		PrintErr(no, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		PrintErr(no, err)
	}

	return client, ctx
}

// ParseGroupMembersCVS parses the supplied CVS group member file
func ParseGroupMembersCVS(file string, port string) ([]string, []string, error) {
	// adapted from https://stackoverflow.com/questions/24999079/reading-csv-file-in-go
	f, err := os.Open(file)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	csvr := csv.NewReader(f)
	clPorts := []string{}
	dbPorts := []string{}

	for {
		row, err := csvr.Read()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return clPorts, dbPorts, nil
		}

		/* Remove own port from results if appropriate */
		if row[0] != port {
			clPorts = append(clPorts, row[0])
			dbPorts = append(dbPorts, row[1])
		}
	}
}

// RPCClient makes an RPC connection to the given port
func RPCClient(no string, port string) *rpc.Client {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+port)
	if err != nil {
		PrintErr(no, err)
	}

	PrintMsg(no, "Connection made to "+port)
	return client
}

// ConnectDriver connects driver to a replica
func ConnectDriver(port string) *rpc.Client {
	var result int
	conn := RPCClient("DRIVER", port)
	err := conn.Call("RPCExt.ConnectReplica", InitArgs{Settings: [2]int{0, 0}, TimeInt: 5000}, &result)
	if err != nil {
		PrintErr("DRIVER", err)
	}
	return conn
}

// Terminate is a command from the driver to terminate a replica
func Terminate(port string, conn *rpc.Client, delay int) {
	time.Sleep(time.Duration(delay) * time.Second)
	var result int
	err := conn.Call("RPCExt.TerminateReplica", RPCExtArgs{}, &result)
	if err != nil {
		PrintErr("DRIVER", err)
	}
	PrintMsg("DRIVER", "Done on "+port)
}

// DownloadCvState gets the current database snapshot for CvRDT
// https://godoc.org/go.mongodb.org/mongo-driver/mongo#Collection.Find
// https://github.com/mongodb/mongo-go-driver
func DownloadCvState(col *mongo.Collection, drop string) []CvRecord {
	var result []CvRecord

	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})
	cursor, err := col.Find(context.TODO(), bson.D{}, opts)
	if err != nil {
		PrintErr("CHECKER", err)
	}
	if err = cursor.All(context.TODO(), &result); err != nil {
		PrintErr("CHECKER", err)
	}
	if drop == "1" {
		col.Drop(context.TODO())
	}
	return result
}

// DownloadCmState gets the current database snapshot for CmRDT
func DownloadCmState(col *mongo.Collection, drop string) []CmRecord {
	var result []CmRecord

	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})
	cursor, err := col.Find(context.TODO(), bson.D{}, opts)
	if err != nil {
		PrintErr("CHECKER", err)
	}
	if err = cursor.All(context.TODO(), &result); err != nil {
		PrintErr("CHECKER", err)
	}
	if drop == "1" {
		col.Drop(context.TODO())
	}
	return result
}

// PrintMsg prints message to console from a replica
func PrintMsg(no string, msg string) {
	if no == "DRIVER" || no == "CHECKER" {
		fmt.Println(no + ": " + msg)
	} else {
		fmt.Println("REPLICA " + no + ": " + msg)
	}
}

// PrintErr prints error to console from a replica and exits
func PrintErr(no string, err error) {
	if no == "DRIVER" || no == "CHECKER" {
		fmt.Println(no+": ", err)
	} else {
		fmt.Println("REPLICA "+no+": ", err)
	}
	os.Exit(1)
}
