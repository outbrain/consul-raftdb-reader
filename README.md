# Consul RAFT database reader

This is a small tool to read Consul's commit log from the RAFT database and output the events in json format.


## Installing and building

```
go get -u github.com/outbrain/consul-raftdb-reader
```

## Using

Copy the `raft/raft.db` file from a Consul server - the Consul server process must be turned off before copying because the file is continuosly mutated. After copying, open the file with this tool:

```
consul-raftdb-reader raft.db > raft.events
```