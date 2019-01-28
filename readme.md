## etcd测试方案

测试原理：
* 利用`goreman`自动运行etcd集群并配置etcd-proxy
* 利用`etcd-test`创建多个watch链接，并以一定的频率对watch的key进行put操作
* 通过`etcd-proxy`支持的api调用的方式模拟不同的网络状况 [官方文档](https://github.com/etcd-io/etcd/blob/master/functional/README.md)
* 通过`Prometheus`从etcd获取metrics并显示成图表, 可以使用`grafana`使图表更易于观察 [官方文档](https://coreos.com/etcd/docs/latest/op-guide/monitoring.html)


### 1. 准备工具

* etcd官方源码
    ```sh
    go get go.etcd.io/etcd
    cd $GOPATH/src/go.etcd.io/etcd
    ./build
    ./functional/build
    ```
* goreman
    ```sh
    go get github.com/mattn/goreman
    ```
* etcd-test
    ```sh
    go get github.com/yoozoo/etcd-test
    ```
* Prometheus
    ```sh
    PROMETHEUS_VERSION="1.3.1"
    wget https://github.com/prometheus/prometheus/releases/download/v$PROMETHEUS_VERSION/prometheus-$PROMETHEUS_VERSION.linux-amd64.tar.gz -O /tmp/prometheus-$PROMETHEUS_VERSION.linux-amd64.tar.gz
    tar -xvzf /tmp/prometheus-$PROMETHEUS_VERSION.linux-amd64.tar.gz --directory /tmp/ --strip-components=1
    /tmp/prometheus -version
    ```
* Grafana(Optinal)
### 2. 配置文件

* goraman对应的Procfile文件在etcd源码目录的 functional/Procfile-proxy, 内容如下：
    ```
    s1: bin/etcd --name s1 --data-dir /tmp/etcd-proxy-data.s1 --listen-client-urls http://127.0.0.1:1379 --advertise-client-urls http://127.0.0.1:13790 --listen-peer-urls http://127.0.0.1:1380 --initial-advertise-peer-urls http://127.0.0.1:13800 --initial-cluster-token tkn --initial-cluster 's1=http://127.0.0.1:13800,s2=http://127.0.0.1:23800,s3=http://127.0.0.1:33800' --initial-cluster-state new

    s1-client-proxy: bin/etcd-proxy --from localhost:13790 --to localhost:1379 --http-port 1378
    s1-peer-proxy: bin/etcd-proxy --from localhost:13800 --to localhost:1380 --http-port 1381

    s2: bin/etcd --name s2 --data-dir /tmp/etcd-proxy-data.s2 --listen-client-urls http://127.0.0.1:2379 --advertise-client-urls http://127.0.0.1:23790 --listen-peer-urls http://127.0.0.1:2380 --initial-advertise-peer-urls http://127.0.0.1:23800 --initial-cluster-token tkn --initial-cluster 's1=http://127.0.0.1:13800,s2=http://127.0.0.1:23800,s3=http://127.0.0.1:33800' --initial-cluster-state new

    s2-client-proxy: bin/etcd-proxy --from localhost:23790 --to localhost:2379 --http-port 2378
    s2-peer-proxy: bin/etcd-proxy --from localhost:23800 --to localhost:2380 --http-port 2381

    s3: bin/etcd --name s3 --data-dir /tmp/etcd-proxy-data.s3 --listen-client-urls http://127.0.0.1:3379 --advertise-client-urls http://127.0.0.1:33790 --listen-peer-urls http://127.0.0.1:3380 --initial-advertise-peer-urls http://127.0.0.1:33800 --initial-cluster-token tkn --initial-cluster 's1=http://127.0.0.1:13800,s2=http://127.0.0.1:23800,s3=http://127.0.0.1:33800' --initial-cluster-state new

    s3-client-proxy: bin/etcd-proxy --from localhost:33790 --to localhost:3379 --http-port 3378
    s3-client-proxy: bin/etcd-proxy --from localhost:33800 --to localhost:3380 --http-port 3381
    ```
    若etcd集群部署在多台机器上，可放弃goreman, 根据以上配置文件在对应机器上手动运行命令。
* 对应上例Procfile的Prometheus配置文件如下
    ```yml
    global:
    scrape_interval: 10s
    scrape_configs:
    - job_name: test-etcd
        static_configs:
        - targets: ['127.0.0.1:1379','127.0.0.1:2379','127.0.0.1:3379']
    ```
* etcd-test只支持flag形式的配置，可以配置的值参考运行时的help
### 3. 运行
```sh
cd $GOPATH/src/go.etcd.io/etcd
goreman -f functional/Procfile-proxy start
```
```sh
prometheus
```
```sh
etcd-test --endpoints='127.0.0.1:1379','127.0.0.1:2379','127.0.0.1:3379'
```
Prometheus默认把结果显示在127.0.0.1:9090/graph

### 4. 模拟网络环境
以上文Procfile文件为例
* 客户端延迟
    ```bash
    $ curl -L http://localhost:2378/delay-tx -X PUT \
    -d "latency=5s&random-variable=100ms"
    # added send latency 5s±100ms (current latency 4.92143955s)

    $ curl -L http://localhost:2378/delay-tx
    # current send latency 4.92143955s

    $ ETCDCTL_API=3 ./bin/etcdctl \
    --endpoints localhost:23790 \
    --command-timeout=3s \
    put foo bar
    # Error: context deadline exceeded

    $ curl -L http://localhost:2378/delay-tx -X DELETE
    # removed latency 4.92143955s

    $ curl -L http://localhost:2378/delay-tx
    # current send latency 0s

    $ ETCDCTL_API=3 ./bin/etcdctl \
    --endpoints localhost:23790 \
    --command-timeout=3s \
    put foo bar
    # OK
    ```

* 暂停向客户端发送

    ```bash
    $ curl -L http://localhost:2378/pause-tx -X PUT
    # paused forwarding [tcp://localhost:23790 -> tcp://localhost:2379]

    $ ETCDCTL_API=3 ./bin/etcdctl --endpoints localhost:23790 put foo bar
    # Error: context deadline exceeded

    $ curl -L http://localhost:2378/pause-tx -X DELETE
    # unpaused forwarding [tcp://localhost:23790 -> tcp://localhost:2379]
    ```

* 丢弃包

    ```bash
    $ curl -L http://localhost:2378/blackhole-tx -X PUT
    # blackholed; dropping packets [tcp://localhost:23790 -> tcp://localhost:2379]

    $ ETCDCTL_API=3 ./bin/etcdctl --endpoints localhost:23790 put foo bar
    # Error: context deadline exceeded

    $ curl -L http://localhost:2378/blackhole-tx -X DELETE
    # unblackholed; restart forwarding [tcp://localhost:23790 -> tcp://localhost:2379]
    ```
