Python indexer for 2 experimental schemas - the one you want is `flat_typed`.

(Type `flat` was implemented just as a stepping stone and reference for performance; can't use it unfortunately).

```
./cli.py --help
usage: cli.py [-h] [--addr ADDR] [--index INDEX] [--schema SCHEMA] {index,create,delete} ...

positional arguments:
  {index,create,delete}
    index               index <num> random devices
    create              create index
    delete              delete index

optional arguments:
  -h, --help            show this help message and exit
  --addr ADDR           elastisearch address (default http://localhost:9200)
  --index INDEX         elastisearch index
  --schema SCHEMA       index schema (flat|flat_typed)
```

Create an index:

```
./cli.py --index flat-devices-tenant1 --schema flat_typed create
```

Seed the index with `--num` random devices:

```
./cli.py --index flat-devices-tenant1 --schema flat_typed index --num 1000000
```

Benchmark (here using `go-wrk`; assumes 9200 is ssh port-forwarded to the benchmark machine):
```
go-wrk -H "content-type:application/json" -M POST -d 5 -c 10 -body @./benchmark/flat_typed/eq.json  http://localhost:9200/flat-devices-tenant2/_search
```
(see other example queries in `./benchmark/flat_typed`)

Delete the index when done:

```
./cli.py --index flat-devices-tenant1 delete
```
