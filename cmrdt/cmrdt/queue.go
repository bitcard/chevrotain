package main

import (
	"errors"
	"fmt"
	"time"

	"../../util"
	"github.com/savreline/GoVector/govec/vclock"
)

// OpCode is an operation code
type OpCode int

// OpCodes
const (
	IK OpCode = iota + 1
	IV
	RK
	RV
)

// OpNode represents a node in the operation wait queue
type OpNode struct {
	Type       OpCode
	Key, Value string
	Timestamp  vclock.VClock
	Pid        string
	ConcOp     bool
}

// translate operation code from string to op code
func lookupOpCode(opName string) OpCode {
	if opName == "IK" {
		return IK
	} else if opName == "IV" {
		return IV
	} else if opName == "RK" {
		return RK
	} else if opName == "RV" {
		return RV
	} else {
		util.PrintErr(noStr, errors.New("lookupOpCode: unknown operation"))
		return 0
	}
}

// Print the Queue
func printQueue() {
	lock.Lock()
	for n := queue.Front(); n != nil; n = n.Next() {
		eLog = eLog + fmt.Sprintln(n.Value)
	}
	eLog = eLog + "\n"
	lock.Unlock()
}

// insert a node into the correct location in the queue
func addToQueue(node OpNode) {
	lock.Lock()
	if queue.Front() == nil {
		queue.PushFront(node)
		lock.Unlock()
		return
	}
	for curNode := queue.Front(); curNode != nil; curNode = curNode.Next() {
		a := node.Timestamp.Compare(curNode.Value.(OpNode).Timestamp, vclock.Ancestor)
		c := node.Timestamp.Compare(curNode.Value.(OpNode).Timestamp, vclock.Concurrent)
		e := node.Timestamp.Compare(curNode.Value.(OpNode).Timestamp, vclock.Equal)

		if c || e {
			node.ConcOp = true
			temp := curNode.Value.(OpNode)
			temp.ConcOp = true
			curNode.Value = temp
			queue.InsertAfter(node, curNode)
			break
		}
		if a {
			queue.InsertBefore(node, curNode)
			break
		}
	}
	lock.Unlock()
}

// process some of the operations that are queued up
func processQueue() {
	for {
		time.Sleep(3 * time.Second)
		lock.Lock()
		eliminateConcOps()
		processQueueHelper()
		lock.Unlock()
	}
}

// does the actual processing of queue operations
func processQueueHelper() {
	// TODO
}

// eliminate concurrent operations from the queue using predefined preference
func eliminateConcOps() {
	// TODO
}
