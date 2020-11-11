package main

/* In this file
1. InsertValue Ext RPC method
2. InsertValueLocal (that works with the local db)
*/

import (
	"context"

	"../../util"
	"github.com/savreline/GoVector/govec"
	"go.mongodb.org/mongo-driver/bson"
)

// InsertValue inserts value into the given key
func (t *RPCExt) InsertValue(args *util.ValueArgs, reply *int) error {
	InsertValueLocal(args.Key, args.Value)
	broadcastInsert(args.Key, args.Value)
	return nil
}

// InsertValueLocal inserts the value into the local db
func InsertValueLocal(key string, value string) {
	logger.LogLocalEvent("Inserting value"+value, govec.GetDefaultLogOptions())
	filter := bson.D{{Key: "name", Value: key}}
	update := bson.D{{Key: "$push", Value: bson.D{
		{Key: "values", Value: value}}}}

	_, err := db.Collection("kvs").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		util.PrintErr(err)
	}
	// fmt.Printf("Matched %v documents and updated %v documents.\n",
	// 	updateResult.MatchedCount, updateResult.ModifiedCount)
	util.PrintMsg(no, "Inserted Value "+value)
}
