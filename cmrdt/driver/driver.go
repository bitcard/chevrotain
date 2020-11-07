package main

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"../../util"
	"github.com/DistributedClocks/GoVector/govec"
)

var ports []string
var logger *govec.GoLog

func main() {
	/* Parse Group Membership */
	var err error
	ports, _, err = util.ParseGroupMembersCVS("ports.csv", "")
	if err != nil {
		util.PrintErr(err)
	}
	noReplicas := len(ports)

	/* Init Cloks */
	logger = govec.InitGoVector("Drv", "Drv", govec.GetDefaultConfig())

	/* Tests */
	for i := 0; i < noReplicas; i++ {
		// go simpleTest(i)
	}
	wikiTest()

	for {
	}
}

// simpleTest
func simpleTest(no int) {
	/* Connect to the Replica and Connect the Replica */
	var result int
	conn := util.RPCClient(logger, ports[no], "DRIVER: ")
	err := conn.Call("RPCExt.ConnectReplica", util.ConnectArgs{}, &result)
	if err != nil {
		util.PrintErr(err)
	}

	/* Inserts */
	for i := 0; i < 10; i++ {
		num := (i + 1) * (no + 1)
		conn.Call("RPCExt.InsertKey", util.KeyArgs{Key: strconv.Itoa(num)}, &result)
	}
}

// wikiTest
func wikiTest() {
	loadPages("Java", 0)
	loadPages("C--", 1)
	loadPages("C++", 2)
}

// https://stackoverflow.com/questions/12518876/how-to-check-if-a-file-exists-in-go
func loadPages(startPage string, no int) {
	pathHead := "../../crawler/" + startPage + "/"
	lastPage := startPage
	maxDepth := 3

	/* Connect to the Replica and Connect the Replica */
	var result int
	conn := util.RPCClient(logger, ports[no], "DRIVER: ")
	err := conn.Call("RPCExt.ConnectReplica", util.ConnectArgs{}, &result)
	if err != nil {
		util.PrintErr(err)
	}

	/* Init Queue */
	var queue []string
	queue = append(queue, startPage)

	/* BFS */
	i := 0
	for len(queue) > 0 && i < maxDepth {
		if queue[0] == lastPage {
			i++
		}

		/* Pop off queue */
		curPage := queue[0]
		queue = queue[1:]

		/* Read the file */
		curPageSp := strings.Replace(curPage, "_", " ", -1)
		body, err := ioutil.ReadFile(pathHead + curPageSp + ".link")
		if err != nil {
			continue
		}

		/* Add the linked files to queue: check to make sure last file actually exists */
		linkedPages := strings.Split(string(body[:]), "\n")
		if curPage == lastPage && i < maxDepth {
			m := 2
			for {
				lastPage = linkedPages[len(linkedPages)-m]
				lastPageSp := strings.Replace(lastPage, "_", " ", -1)
				if _, err := os.Stat(pathHead + lastPageSp + ".link"); err == nil {
					break
				}
				m++
			}
		}

		/* Insert Key */
		err = conn.Call("RPCExt.InsertKey", util.KeyArgs{Key: curPage}, &result)
		if err != nil {
			util.PrintErr(err)
		}

		/* Add to Queue and Insert Value */
		for j := 0; j < len(linkedPages)-1; j++ {
			queue = append(queue, linkedPages[j])

			/* Insert Value */
			err = conn.Call("RPCExt.InsertValue", util.ValueArgs{Key: curPage, Value: linkedPages[j]}, &result)
			if err != nil {
				util.PrintErr(err)
			}
		}
	}
}
