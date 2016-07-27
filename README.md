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

## Indexing with ELK

A common use case is analyzing the dump with ELK. You can use the included ElasticSearch [index template](raft.es.template.json) to get ElasticSearch to index the fields properly.

```
curl -XPUT http://localhost:9200/_template/consul_raft -d @raft.es.template.es
```

After the index template has been loaded, index the dump to ElasticSearch with Logstash:
```
consul-raftdb-reader raft.db | logstash agent -e 'input { stdin { codec => json_lines } } output { elasticsearch { index => "consul_raft" document_type => "raft_txn" hosts => ["localhost:9200"] } }'
```

Now you can analyze the dump in Kibana.
