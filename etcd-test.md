# etcd cluster test result

## tools used

* benchmark from github.com/etcd-io/etcd/tree/master/tools/benchmark , the benchmark tool from etcd
* toxiproxy from github.com/Shopify/toxiproxy , a tool to test with varying network condition

## test env

* etcd cluster: 5 node, bandwidth between node 100MB. latency between node about 10M.
* test machine: 1 vm. latency to the node are : 10ms, 10ms, 0.2ms, 8ms, 8ms

## system capacity test result

### write performance

example test command is :

```bash
./benchmark put --endpoints "10.22.51.37:2379,10.22.51.38:2379,10.18.32.61:2379,10.6.48.36:2379,10.6.48.44:2379" --conns=400  --clients=400 --val-size 1900 --total=3000000 --rate 0 --key-size=100 --key-space-size=1000 --compact-index-delta 1000 --compact-interval 10s

```

test result

```txt
(key size, value size) = (100, 100)
    Slowest:      0.2442 secs.
    Fastest:      0.0115 secs.
    Average:      0.0287 secs.
    Stddev:       0.0106 secs.
    Requests/sec: 15436.8672

bandwidth used for this request : etcd->client 3.5m, etcd->etcd 15M


(key size, value size) = (100, 1900)
    Slowest:      0.8141 secs.
    Fastest:      0.0101 secs.
    Average:      0.0429 secs.
    Stddev:       0.0668 secs.
    Requests/sec: 9843.6390

bandwidth used for this request : etcd->client 19.2M, etcd->etcd 80M
```

### read performance

#### setup

* delete all keys from etcd

```bash
   /etcdctl del "" --prefix --endpoints "10.22.51.37:2379" )
```

* create 100k kv in etcd with (key size, value size) = (100, 100)

```bash
./benchmark put --endpoints "10.22.51.37:2379,10.22.51.38:2379,10.18.32.61:2379,10.6.48.36:2379,10.6.48.44:2379" --conns=4000  --clients=800 --val-size 100 --total=100000 --rate 0 --key-size=100 --key-space-size=100000 --compact-index-delta 10000 --compact-interval 1m --sequential-keys
```

* create 1000 custom keys

```bash
for x in {1..1000} ; do ./etcdctl put "testkey$x" "testkey$x" --endpoints "10.22.51.38:2379" ;done
```

#### read all keys

example test command is

```bash
./benchmark range $'\001' $'\377' --rate  0 --total 1000 --clients 10 --conns 10 --endpoints "10.22.51.37:2379,10.22.51.38:2379,10.18.32.61:2379,10.6.48.36:2379,10.6.48.44:2379"
```

test result

```txt
with 10 clients
  Slowest:      4.4398 secs.
  Fastest:      0.6753 secs.
  Average:      1.9217 secs.
  Stddev:       0.8050 secs.
  Requests/sec: 4.8435

with 5 clients
  Slowest:      3.7799 secs.
  Fastest:      0.5665 secs.
  Average:      1.1838 secs.
  Stddev:       0.4057 secs.
  Requests/sec: 4.7529

with 5 clients from 1 endpoint
   Slowest:      2.6727 secs.
   Fastest:      0.7497 secs.
   Average:      1.2091 secs.
   Stddev:       0.2463 secs.
   Requests/sec: 4.4376

all the read saturated the bandwidth of the test machine. the time for each requestis dominated by network transfer.
```

#### read 100 keys

example test command is

```bash
./benchmark range 'testkey200' 'testkey299' --rate  0 --total 1000000 --clients 1000 --conns 1000 --endpoints "10.22.51.37:2379,10.22.51.38:2379,10.18.32.61:2379,10.6.48.36:2379,10.6.48.44:2379"
```

test result

```txt
1000 client on 5 endpoint
  Slowest:      0.3447 secs.
  Fastest:      0.0090 secs.
  Average:      0.0698 secs.
  Stddev:       0.0347 secs.
  Requests/sec: 14105.8524

  test machine bandwidth usage is around 30M/s

1000 client on 1 endpoint (latency 10ms)
  Slowest:      0.8801 secs.
  Fastest:      0.0305 secs.
  Average:      0.3957 secs.
  Stddev:       0.0728 secs.
  Requests/sec: 2525.5984

```

#### read 10 key

test result

```txt
1000 client on 5 endpoint
  Slowest:      0.1089 secs.
  Fastest:      0.0095 secs.
  Average:      0.0418 secs.
  Stddev:       0.0149 secs.
  Requests/sec: 23604.6647

1000 client on 1 endpoint , 10 ms latency
  Slowest:      0.4850 secs.
  Fastest:      0.0253 secs.
  Average:      0.1176 secs.
  Stddev:       0.0718 secs.
  Requests/sec: 8340.4115

1000 client on 1 endpoint, 0.2 ms
  Slowest:      0.3058 secs.
  Fastest:      0.0089 secs.
  Average:      0.0735 secs.
  Stddev:       0.0451 secs.
  Requests/sec: 13522.9060

for small read, the read speed is dominated by the network latency between tclient and the node used.
```

### watch performance

#### 1000 key set (application), 10 watcher per key set  (our assumed use case)

example test command

```bash
./benchmark watch --key-size 1000 --key-space-size 10000 --put-rate 1000 --put-total  10000 --streams  1000 --watch-per-stream 10 --watched-key-total 1000   --clients 10000  --conns 2000 --endpoints  "10.22.51.37:2379,10.22.51.38:2379,10.18.32.61:2379,10.6.48.36:2379,10.6.48.44:2379"
```

test result

```txt
create watcher:
    Slowest:      0.0985 secs.
    Fastest:      0.0005 secs.
    Average:      0.0157 secs.
    Stddev:       0.0136 secs.

10000 key change with (key size) = (100)
    Slowest:      0.0001 secs.
    Fastest:      0.0000 secs.
    Average:      0.0000 secs.
    Stddev:       0.0000 secs.

10000 key change with (key size) = (1000)
    Slowest:      0.0001 secs.
    Fastest:      0.0000 secs.
    Average:      0.0000 secs.
    Stddev:       0.0000 secs.

the time between key set to watch event received could be small as the two event (set key and key change detected) can happen at same time on the server side
```

#### 10 key set (application), 1000 watcher per key set

test result

```txt
10000 key change with (key size) = (1000)
  Slowest:      0.0272 secs.
  Fastest:      0.0000 secs.
  Average:      0.0001 secs.
  Stddev:       0.0003 secs.

there are slight delay when the number of watcher is large on the same key. this most likely is due to one notify for every watcher and large number of watchers will naturally require a lot more notify messages.
```

#### 100K connection with 100k watcher, 1 watcher per key

test result

```txt
create watcher
    Slowest:      0.1617 secs.
    Fastest:      0.0048 secs.
    Average:      0.0373 secs.
    Stddev:       0.0220 secs.

watch result is about same as in 10 watcher per key. so etcd are able support large number of watchers, and number of watcher does not affect the performance of notify.
```

### test with toxiproxy

#### toxiproxy setup

toxiproxy is a tool that can let you control network parameters like latency, bandwidth,jitter etc. do take note that the parameters (toxic in toxiproxy) is per connection. so in our test using toxiproxy, we need to change the connection(--conns) to 1.

* run toxiproxy-server.

```bash
./toxiproxy-serer > a.txt &
```

* setup a proxy

```bash
./toxiproxy-cli create test -l localhost:2379 -u 10.22.51.37:2379
```

we will have the proxy forward localhost:2379 to 10.22.51.37:2379 . the test follows are all using single endpiont localhost:2379

#### test with added latency

command to change latency is

```bash
#create latency toxic (90ms delay on upstream)
./toxiproxy-cli toxic add test -n test1 -t latency -a latency=90 -u
#update latency toxic (change to 240ms delay on upstream)
./toxiproxy-cli toxic update test -n test1 -a latency=240 -u

```

* 100 ms latency ()

```txt
to read 10 keys with 1000 client
  Slowest:      0.2757 secs.
  Fastest:      0.1199 secs.
  Average:      0.1930 secs.
  Stddev:       0.0142 secs.
  Requests/sec: 5084.1317
```

* 250 ms

```txt
to read 10 keys with 1000 client
  Slowest:      0.5582 secs.
  Fastest:      0.2899 secs.
  Average:      0.4884 secs.
  Stddev:       0.0190 secs.
  Requests/sec: 2030.4342
```

we can see that latency affects the performance great request.

#### test with limited bandwidth

command to limit bandwidth is

```bash
#downstream/upstream limit to 1MB
./toxiproxy-cli toxic add test -n test2 -t bandwidth -a rate=1024 -d
./toxiproxy-cli toxic add test -n test2 -t bandwidth -a rate=1024 -u
```

test result

```txt
write with (key size, value size) = (100, 1900)
  Slowest:      1.4995 secs.
  Fastest:      0.0503 secs.
  Average:      0.7909 secs.
  Stddev:       0.1151 secs.
  Requests/sec: 502.6996

read 100 keys with 1000 client
  Slowest:      7.3925 secs.
  Fastest:      0.0243 secs.
  Average:      3.9924 secs.
  Stddev:       0.5538 secs.
  Requests/sec: 246.3672

read 10 keys with 1000 client
  Slowest:      0.6070 secs.
  Fastest:      0.0148 secs.
  Average:      0.3958 secs.
  Stddev:       0.0403 secs.
  Requests/sec: 2485.2856

```

#### test with limited bandwidth and high latency

this is tested with 1MB bandwidth and 500ms latency

test result

```txt
read 100 keys with 1000 client
  Slowest:      10.0845 secs.
  Fastest:      0.5232 secs.
  Average:      4.0866 secs.
  Stddev:       0.3848 secs.
  Requests/sec: 240.7642

read 10 keys with 1000 client
  Slowest:      3.3011 secs.
  Fastest:      0.5111 secs.
  Average:      0.5968 secs.
  Stddev:       0.2955 secs.
  Requests/sec: 1651.8431

```

with large latency, the system capacity is actually dropped for small read request.

