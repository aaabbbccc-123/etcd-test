# etcd cluster 测试结果

## 使用的工具

* benchmark  ( github.com/etcd-io/etcd/tree/master/tools/benchmark ), etcd 本身的性能测试工具
* toxiproxy （ github.com/Shopify/toxiproxy ), 一个测试应用对不同网络环境反应的工具

## 测试环境

* etcd 集群: 5个节点, 节点之间带宽大约100MB. 节点之间的延迟大约 10毫秒
* 测试机器: 1个虚拟机. 到各节点的延迟为 : 10ms, 10ms, 0.2ms, 8ms, 8ms 。测试开始之间、前要改可打开文件到 102400 ( ulimit -n 102400 )
## 系统容量测试结果

### 写入性能

使用的测试命令举例 :

```bash
./benchmark put --endpoints "10.22.51.37:2379,10.22.51.38:2379,10.18.32.61:2379,10.6.48.36:2379,10.6.48.44:2379" --conns=400  --clients=400 --val-size 1900 --total=3000000 --rate 0 --key-size=100 --key-space-size=1000 --compact-index-delta 1000 --compact-interval 10s

```

测试结果

```txt
(键长， 值长) = (100, 100)
    Slowest:      0.2442 secs.
    Fastest:      0.0115 secs.
    Average:      0.0287 secs.
    Stddev:       0.0106 secs.
    Requests/sec: 15436.8672

过程中使用的带宽 : etcd->client 3.5MB, etcd->etcd 15MB


(键长， 值长) = (100, 1900)
    Slowest:      0.8141 secs.
    Fastest:      0.0101 secs.
    Average:      0.0429 secs.
    Stddev:       0.0668 secs.
    Requests/sec: 9843.6390

过程中使用的带宽 : etcd->client 19.2M, etcd->etcd 80M
```

### 读取性能

#### 建立读取环境

* 删除集群中所有的键

```bash
   /etcdctl del "" --prefix --endpoints "10.22.51.37:2379" )
```

* 建立10万个kv (键长， 值长)= (100, 100)

```bash
./benchmark put --endpoints "10.22.51.37:2379,10.22.51.38:2379,10.18.32.61:2379,10.6.48.36:2379,10.6.48.44:2379" --conns=4000  --clients=800 --val-size 100 --total=100000 --rate 0 --key-size=100 --key-space-size=100000 --compact-index-delta 10000 --compact-interval 1m --sequential-keys
```

* 建立1000个 自定义的kv

```bash
for x in {1..1000} ; do ./etcdctl put "testkey$x" "testkey$x" --endpoints "10.22.51.38:2379" ;done
```

#### 读取所有键

使用的测试命令举例

```bash
./benchmark range $'\001' $'\377' --rate  0 --total 1000 --clients 10 --conns 10 --endpoints "10.22.51.37:2379,10.22.51.38:2379,10.18.32.61:2379,10.6.48.36:2379,10.6.48.44:2379"
```

测试结果

```txt
10个客户端
  Slowest:      4.4398 secs.
  Fastest:      0.6753 secs.
  Average:      1.9217 secs.
  Stddev:       0.8050 secs.
  Requests/sec: 4.8435

5个客户端
  Slowest:      3.7799 secs.
  Fastest:      0.5665 secs.
  Average:      1.1838 secs.
  Stddev:       0.4057 secs.
  Requests/sec: 4.7529

5个客户端，只从一个节点读取
   Slowest:      2.6727 secs.
   Fastest:      0.7497 secs.
   Average:      1.2091 secs.
   Stddev:       0.2463 secs.
   Requests/sec: 4.4376

所有的读取都受到了带宽的限制. 读取时间最主要的部分是网络传输的时间
```

#### 读取100个键

使用的测试命令举例

```bash
./benchmark range 'testkey200' 'testkey299' --rate  0 --total 1000000 --clients 1000 --conns 1000 --endpoints "10.22.51.37:2379,10.22.51.38:2379,10.18.32.61:2379,10.6.48.36:2379,10.6.48.44:2379"
```

测试结果

```txt
1000个客户端 使用5个etcd节点
  Slowest:      0.3447 secs.
  Fastest:      0.0090 secs.
  Average:      0.0698 secs.
  Stddev:       0.0347 secs.
  Requests/sec: 14105.8524

  测试机的带宽使用大约在 30MB/s

1000个客户端，只使用一个节点 (延迟大约 10ms)
  Slowest:      0.8801 secs.
  Fastest:      0.0305 secs.
  Average:      0.3957 secs.
  Stddev:       0.0728 secs.
  Requests/sec: 2525.5984

```

#### 读取10个键

测试结果

```txt
1000个客户端， 5个节点
  Slowest:      0.1089 secs.
  Fastest:      0.0095 secs.
  Average:      0.0418 secs.
  Stddev:       0.0149 secs.
  Requests/sec: 23604.6647

1000个客户端， 使用1个etcd节点，10ms 延迟
  Slowest:      0.4850 secs.
  Fastest:      0.0253 secs.
  Average:      0.1176 secs.
  Stddev:       0.0718 secs.
  Requests/sec: 8340.4115

1000个客户端， 使用1个etcd节点， 0.2ms 延迟
  Slowest:      0.3058 secs.
  Fastest:      0.0089 secs.
  Average:      0.0735 secs.
  Stddev:       0.0451 secs.
  Requests/sec: 13522.9060

对于很小的读取操作, 速度主要是受客户端和使用的节点间网络延迟的影响.
```

### 更改监控性能

#### 1000个监控对象 (应用), 每个监控对象 10个 观察者  (假设的使用情况)

使用的测试命令举例

```bash
./benchmark watch --key-size 1000 --key-space-size 10000 --put-rate 1000 --put-total  10000 --streams  1000 --watch-per-stream 10 --watched-key-total 1000   --clients 10000  --conns 2000 --endpoints  "10.22.51.37:2379,10.22.51.38:2379,10.18.32.61:2379,10.6.48.36:2379,10.6.48.44:2379"
```

测试结果

```txt
创建观察者:
    Slowest:      0.0985 secs.
    Fastest:      0.0005 secs.
    Average:      0.0157 secs.
    Stddev:       0.0136 secs.

10000个键值更改 (键长) = (100)
    Slowest:      0.0001 secs.
    Fastest:      0.0000 secs.
    Average:      0.0000 secs.
    Stddev:       0.0000 secs.

10000个键值更改 (键长) = (1000)
    Slowest:      0.0001 secs.
    Fastest:      0.0000 secs.
    Average:      0.0000 secs.
    Stddev:       0.0000 secs.

从客户端拿到键值改变成功 到收到 键值更改 的消息 的时间间隔 会非常小。这是因为 键值更改 的消息是在服务器端 更改成功后创建发送的。 这个时间和更改成功的返回消息发送时间差别很微小。
```

#### 10 个监控对象 (应用), 每个监控对象 1000 个观察者

测试结果

```txt
10000个键值更改 (键长) = (1000)
  Slowest:      0.0272 secs.
  Fastest:      0.0000 secs.
  Average:      0.0001 secs.
  Stddev:       0.0003 secs.

这个测试里收到键值更改的提醒消息相对慢一些。这个主要应该是因为这个测试中一个键的改变会创建并发送相对多很多的 提醒消息。消息的数量增加需要更长的传输时间。
```

#### 10万个链接， 10万个观察者， 每个键一个观察者

测试结果

```txt
创建观察者:
    Slowest:      0.1617 secs.
    Fastest:      0.0048 secs.
    Average:      0.0373 secs.
    Stddev:       0.0220 secs.

100000 个键值更改
    Slowest:      0.0001 secs.
    Fastest:      0.0000 secs.
    Average:      0.0000 secs.
    Stddev:       0.0000 secs.

测试结果和1000个观察者的情况区别不大. etcd对大量观察者支持的非常好，更多的观察者存在并不会影响到提醒的性能。
```

## 使用toxiproxy的测试

### toxiproxy 设置

toxiproxy 是一个能让你管理一个连接的各种参数（比如延迟，带宽）的工具。需要注意的是这些参数（toxic）都是对应一个连接的. 在我们使用toxiproxy做的测试中，需要只使用一个tcp连接（--conns 1）。

* 运行 toxiproxy 代理 服务.

```bash
./toxiproxy-serer > a.txt &
```

* 建立一个代理端口

```bash
./toxiproxy-cli create test -l localhost:2379 -u 10.22.51.37:2379
```

这里我们建立了一个从 localhost:2379 到  10.22.51.37:2379 的端口代理. 下面的测试都是使用的一个节点 localhost:2379

### test with added latency

更改延迟的命令 (节点本身有10ms延迟)

```bash
#创建一个upstream上90ms的延迟 (90ms delay on upstream)
./toxiproxy-cli toxic add test -n test1 -t latency -a latency=90 -u
#更新延迟到240ms
./toxiproxy-cli toxic update test -n test1 -a latency=240 -u

```

* 100ms延迟下的读取

```txt
1000个客户端， 读取10个键值
  Slowest:      0.2757 secs.
  Fastest:      0.1199 secs.
  Average:      0.1930 secs.
  Stddev:       0.0142 secs.
  Requests/sec: 5084.1317
```

* 250ms延迟下的读取

```txt
1000个客户端， 读取10个键值
  Slowest:      0.5582 secs.
  Fastest:      0.2899 secs.
  Average:      0.4884 secs.
  Stddev:       0.0190 secs.
  Requests/sec: 2030.4342
```

能看到对于很小的读取操作，延迟对性能影响很大.

### 限制带宽下的测试

限制带宽的命令

```bash
#把downstream/upstream 带宽限制到1MB
./toxiproxy-cli toxic add test -n test2 -t bandwidth -a rate=1024 -d
./toxiproxy-cli toxic add test -n test3 -t bandwidth -a rate=1024 -u
```

测试结果

```txt
写入 (键长，值长) = (100, 1900)
  Slowest:      1.4995 secs.
  Fastest:      0.0503 secs.
  Average:      0.7909 secs.
  Stddev:       0.1151 secs.
  Requests/sec: 502.6996

1000客户端， 读取100 个键
  Slowest:      7.3925 secs.
  Fastest:      0.0243 secs.
  Average:      3.9924 secs.
  Stddev:       0.5538 secs.
  Requests/sec: 246.3672

1000客户端， 读取10 个键
  Slowest:      0.6070 secs.
  Fastest:      0.0148 secs.
  Average:      0.3958 secs.
  Stddev:       0.0403 secs.
  Requests/sec: 2485.2856

```

### 限制带宽同时高延迟的测试

这个测试的参数是 500ms延迟，1MB带宽

测试结果

```txt
1000客户端， 读取100 个键
  Slowest:      10.0845 secs.
  Fastest:      0.5232 secs.
  Average:      4.0866 secs.
  Stddev:       0.3848 secs.
  Requests/sec: 240.7642

1000客户端， 读取10 个键
  Slowest:      3.3011 secs.
  Fastest:      0.5111 secs.
  Average:      0.5968 secs.
  Stddev:       0.2955 secs.
  Requests/sec: 1651.8431

```

很明显的高延迟下，数据量小的操作 性能差了很多

## 没测试的情况

没有跑任何需要改变节点间网络环境的测试。这样的测试需要能管理网络硬件的工具（比如虚拟机里的网络设备）。
