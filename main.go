package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/go-msgpack/codec"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
	"gopkg.in/alecthomas/kingpin.v2"
	"reflect"
)

var (
	raftDBFile = kingpin.Arg("path", "RAFT data dir").ExistingFile()
	mh         codec.MsgpackHandle
)

// https://github.com/hashicorp/raft/blob/master/log.go#L6
var logTypes = []string{
	"Command",
	"Noop",
	"AddPeer",
	"RemovePeer",
	"Barrier",
}

func main() {
	kingpin.Parse()
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))
	raftStore, err := raftboltdb.NewBoltStore(*raftDBFile)
	if err != nil {
		kingpin.Fatalf("Failed to open RAFT store file %s\nError: %s\n", *raftDBFile, err.Error())
	}

	var start, last uint64

	start, err = raftStore.FirstIndex()
	if err != nil {
		kingpin.Fatalf("Error reading from RAFT store: %s\n", err.Error())
	}

	last, err = raftStore.LastIndex()
	if err != nil {
		kingpin.Fatalf("Error reading from RAFT store: %s\n", err.Error())
	}

	kingpin.Errorf("First index: %d, Last index: %d, reading %d events\n", start, last, last-start)

	var log raft.Log

	for i := start; i < last; i++ {
		raftStore.GetLog(i, &log)
		switch log.Type {
		case 0:
			handleLogCommand(log)
		case 1:
			printJson(log, "", log.Data)
		case 2, 3:
			printJson(log, "", struct{ Peers []string }{Peers: decodePeerMsg(log.Data)})
		default:
			printJson(log, "", log.Data)
		}
	}
}

func handleLogCommand(log raft.Log) {
	buf := log.Data
	if len(buf) > 0 {
		msgType := structs.MessageType(buf[0])
		switch msgType {
		case structs.RegisterRequestType:
			var data structs.RegisterRequest
			printData(log, buf[1:], &data)
		case structs.DeregisterRequestType:
			var data structs.DeregisterRequest
			printData(log, buf[1:], &data)
		case structs.KVSRequestType:
			var data structs.KVSRequest
			printData(log, buf[1:], &data)
		case structs.SessionRequestType:
			var data structs.SessionRequest
			printData(log, buf[1:], &data)
		case structs.ACLRequestType:
			var data structs.ACLRequest
			printData(log, buf[1:], &data)
		case structs.TombstoneRequestType:
			var data structs.TombstoneRequest
			printData(log, buf[1:], &data)
		case structs.CoordinateBatchUpdateType, 134:
			var data structs.Coordinates
			printData(log, buf[1:], &data)
		case structs.PreparedQueryRequestType:
			var data structs.PreparedQueryRequest
			printData(log, buf[1:], &data)
		default:
			kingpin.Errorf("Unknown msg type %d", msgType)
			printMsgPackData(log, reflect.Indirect(reflect.ValueOf(msgType)).Type().Name(), buf[1:])
		}

	} else {
		printJson(log, "", log.Data)
	}
}

func decodePeerMsg(buf []byte) []string {
	var data []string
	if err := codec.NewDecoder(bytes.NewReader(buf), &mh).Decode(&data); err != nil {
		kingpin.Errorf("Error while decoding (generic msgpack) message: %s\n", err.Error())
	}
	return data
}

func printMsgPackData(log raft.Log, msgtype string, buf []byte) {
	var data interface{}
	if err := codec.NewDecoder(bytes.NewReader(buf), &mh).Decode(&data); err != nil {
		kingpin.Errorf("Error while decoding (generic msgpack) message: %s\n", err.Error())
	}
	printJson(log, msgtype, data)
}

func decode(buf []byte, out interface{}) {
	if err := structs.Decode(buf, out); err != nil {
		kingpin.Errorf("Error while decoding message: %s\n", err.Error())
	}
}

func printData(log raft.Log, buf []byte, out interface{}) {
	decode(buf, out)
	n := reflect.Indirect(reflect.ValueOf(out))
	name := n.Type().Name()
	printJson(log, name, out)
}

func printJson(log raft.Log, msgtype string, data interface{}) {
	output := map[string]interface{}{"index": log.Index, "term": log.Term, "type": logTypes[log.Type], "msgtype": msgtype, "data": data}
	if b, err := json.Marshal(output); err != nil {
		kingpin.Errorf("Error serializing data: %s\n", err.Error())
	} else {
		fmt.Println(string(b))
	}
}
