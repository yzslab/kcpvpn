# KCPVPN
[KCP](https://github.com/skywind3000/kcp) based VPN

License: [Apache License, Version 2.0](https://github.com/yzsme/kcpvpn/blob/master/LICENSE)

**本程序仅用于异地组网，不支持其它用途，使用时请严格遵守当地法律**
# 安装
## 方式一：下载Release
见本项目的Release页
## 方式二：自行编译
参考：[https://github.com/yzsme/kcpvpn/blob/master/build.bash](https://github.com/yzsme/kcpvpn/blob/master/build.bash)
```shell script
git clone https://github.com/yzsme/kcpvpn
pushd kcpvpn
bash ./build.bash
ls -l kcpvpn-build
popd
```
# 使用
## 允许以非root用户运行
程序使用到tun/tap，默认情况下仅有root用户可调用相关API，设置以下capabilities后通过非root用户启动程序可监听<1024的端口且可调用tun/tap相关API：
```shell script
setcap cap_net_admin,cap_net_bind_service,cap_net_raw+ep /usr/bin/kcpvpn
```
## 服务器模式
```shell script
user1@server1:~$ kcpvpn server --help
...
   --ip value                       服务器模式下监听的IP
   --port value                     服务器模式下监听的端口
   --udp-mtu value                  UDP数据包MTU (default: 1350)
   --kcp-mode value                 KCP模式: fast3, fast2, fast, normal, manual (default: "fast"，参见KCPTUN)
   --sndwnd value                   set send window size(num of packets) (default: 1024)
   --rcvwnd value                   set receive window size(num of packets) (default: 1024)
   --datashard value, --ds value    set reed-solomon erasure coding - datashard (default: 10)
   --parityshard value, --ps value  set reed-solomon erasure coding - parityshard (default: 3)
   --rapid-fec                      增加少量流量消耗，提升低pps情况下fec的效果，从而降低丢包对延迟的影响
   --dscp value                     set DSCP(6bit) (default: 0)
   --sockbuf value                  per-socket buffer in bytes (default: 4194304)
   --smuxbuf value                  the overall de-mux buffer in bytes (default: 4194304)
   --keepalive value                seconds between heartbeats (default: 10)
   --secret value                   密钥 [$KCPVPN_KEY]
   --crypt value                    加密算法：aes, aes-128, aes-192, salsa20, blowfish, twofish, cast5, 3des, tea, xtea, xor, sm4, none (default: "aes")
   --local-ip value                 本地IP，服务器仅在TUN模式下适用
   --netmask value                  子网掩码，仅TAP模式下适用
   --vni-mtu value                  tun/tap虚拟网卡MTU (default: 1500)
   --full-frame-mtu value           decide automatically if is 0 (default: 0)
   --vni-name-prefix value          服务器模式下虚拟网卡名称前缀 (default: "kvs")
   --vni-mode value                 虚拟网卡模式，tun或tap
   --assignable-ips value           用于分配给客户端的IP范围，格式如：192.168.0.0/24 or 192.168.0.0-192.168.0.255
   --hook-dir value                 hook存放文件夹
   --bridge value                   自动把虚拟网卡加入到所指定名称的桥，仅TAP模式适用
   --tcp                            raw socket模拟TCP模式
```
## 客户端模式
```shell script
user1@client1:~$ kcpvpn client --help
...
   --ip value                       服务器IP
   --port value                     服务器端口
   --udp-mtu value                  UDP数据包MTU (default: 1350)
   --kcp-mode value                 KCP模式: fast3, fast2, fast, normal, manual (default: "fast"，参见KCPTUN)
   --sndwnd value                   set send window size(num of packets) (default: 1024)
   --rcvwnd value                   set receive window size(num of packets) (default: 1024)
   --datashard value, --ds value    set reed-solomon erasure coding - datashard (default: 10)
   --parityshard value, --ps value  set reed-solomon erasure coding - parityshard (default: 3)
   --rapid-fec                      增加少量流量消耗，提升低pps情况下fec的效果，从而降低丢包对延迟的影响
   --dscp value                     set DSCP(6bit) (default: 0)
   --sockbuf value                  per-socket buffer in bytes (default: 4194304)
   --smuxbuf value                  the overall de-mux buffer in bytes (default: 4194304)
   --keepalive value                seconds between heartbeats (default: 10)
   --secret value                   密钥 [$KCPVPN_KEY]
   --crypt value                    加密算法：aes, aes-128, aes-192, salsa20, blowfish, twofish, cast5, 3des, tea, xtea, xor, sm4, none (default: "aes")
   --local-ip value                 指定使用的本地IP，客户端在tun和tap模式下均适用
   --vni-mtu value                  tun/tap虚拟网卡MTU (default: 1500)
   --client-id value                客户端ID，用以服务器根据客户端ID调用Hook，最大长度不可超过16字节
   --vni-name value                 tun/tap网卡名称
   --persistent-vni                 持久型tun/tap网卡，即程序退出后，仍然保留
   --auto-reconnect                 与服务器链接中断后，自动重连
   --bridge value                   自动把虚拟网卡加入到所指定名称的桥，仅TAP模式适用
   --tcp                            raw socket模拟TCP模式
   --on-connected-hook value        客户端与服务器连接成功后调用的hook
...
```
## 使用例子
### 服务器
#### tun模式
如果--local-ip在--assignable-ips范围中，程序会自动排除
```shell script
user1@server1:~$ KCPVPN_KEY="pre-shared-key" kcpvpn server \
    --ip 0.0.0.0 \
    --port 1235 \
    --kcp-mode normal \
    --dscp 46 \
    --local-ip 192.168.88.0 \
    --vni-mode tun \
    --assignable-ips 192.168.88.0/24 \
    --hook-dir /kcpvpn-hooks
```
#### tap模式
tap模式需要注意分配的IP范围需要手动排除network与broadcast地址，并且指定netmask；一般使用tap模式会有桥接的需求，可用使用--bridge指定一个现有的桥，程序将会自动把创建的tap网卡加入到桥
```shell script
user1@server1:~$ KCPVPN_KEY="pre-shared-key" kcpvpn server \
    --ip 0.0.0.0 \
    --port 1235 \
    --kcp-mode normal \
    --dscp 46 \
    --local-ip 192.168.88.1 \
    --vni-mode tap \
    --assignable-ips 192.168.88.2-192.168.88.254 \
    --netmask 255.255.255.0 \
    --hook-dir /kcpvpn-hooks \
    --bridge br0
```
### 客户端
除了KCP协议本身的参数外，客户端无需指定VPN的参数，例如VPN模式（tun或tap），MTU等，这些参数都可以通过握手获取
```shell script
user1@client1:~$ KCPVPN_KEY="pre-shared-key" kcpvpn client \
    --client-id client1 \
    --ip server1 \
    --port 1235 \
    --kcp-mode normal \
    --dscp 46
```
如客户端指定使用固定IP，并且指定期望的MTU值（最终会取服务器与客户端中的最小值）：
```shell script
user1@client1:~$ KCPVPN_KEY="pre-shared-key" kcpvpn client \
    --client-id client1 \
    --ip server1 \
    --port 1235 \
    --kcp-mode normal \
    --dscp 46 \
    --vni-mtu 1200 \
    --local-ip 192.168.88.255 \
    --vni-name kvc_c1
```
## 随系统启动
可参考以下systemd的配置文件
### 服务器
```shell script
user1@server1:~$ cat /etc/systemd/system/kcpvpn-server.service
[Unit]
Description=KCPVPN Server
After=network-online.target

[Service]
Environment="KCPVPN_KEY=pre-shared-key"
Type=simple
User=nobody
ExecStart=/usr/bin/kcpvpn server --ip 0.0.0.0 --port 1235 --udp-mtu 1480 --kcp-mode fast3 --crypt xor --vni-mode tun --vni-mtu 1400 --local-ip 192.168.88.0 --assignable-ips 192.168.88.0/24
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
user1@server1:~$ systemctl daemon-reload
user1@server1:~$ systemctl enable kcpvpn-server
user1@server1:~$ systemctl start kcpvpn-server
```
### 客户端
```shell script
user1@client1:~$ cat /etc/systemd/system/kcpvpn-client.service
[Unit]
Description=KCPVPN Server
After=network-online.target

[Service]
Environment="KCPVPN_KEY=pre-shared-key"
Type=simple
User=nobody
ExecStart=/usr/bin/kcpvpn server --ip server1 --port 1235 --udp-mtu 1480 --kcp-mode fast3 --crypt xor --persistent-vni --auto-reconnect
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
user1@client1:~$ systemctl daemon-reload
user1@client1:~$ systemctl enable kcpvpn-client
user1@client1:~$ systemctl start kcpvpn-client
```
## Hook
在服务器模式下，可通过--hook-dir指定hook存放的目录；客户端模式可通过--on-connected-hook指定连接成功后的hook；Hook可以是Script，也可以是二进制可执行程序。

当客户端与服务器握手成功后，服务器会检查客户端是否有传递client id，如果有，检查目录下是否存在文件on_CLIENTID_connected文件，存在则运行（例如客户端传递的Client ID为client1，则hook文件名为on_client1_connected）。

若文件不存在或者客户端没有传递client id，则检查文件on_connected是否存在，存在则调用

服务器通过以下环境变量给Hook程序传递信息：
* KV_CLIENT_ID: 客户端ID
* KV_VNI_INTERFACE_NAME: 客户端的虚拟网卡名称
* KV_CLIENT_IP: 客户端使用的IP，注意如果是0.0.0.0表示没有给客户端分配IP，客户端也没有指定使用特定IP
* KV_CLIENT_IP_MODE：客户端的IP分配模式
* KV_REMOTE_ADDR：客户端连接服务器使用的IP地址与端口
## tap模式桥接
tap模式附带以太网头，可直接通过桥接进行异地组网。

### 服务器配置示例如下
```shell script
# 客户端连接后，服务器会出现kvs（可通过--vni-name-prefix更改）开头的TAP网卡
user1@server1:~$ ip link show
...
8: kvs0: <BROADCAST,UP,LOWER_UP> mtu 1500 qdisc fq_codel state UNKNOWN mode DEFAULT group default qlen 1000
    link/ether de:0b:f7:84:b5:a9 brd ff:ff:ff:ff:ff:ff
...

# 创建桥接网卡
user1@server1:~$ brctl addbr kvsbr0
user1@server1:~$ ip link set kvsbr0 up
user1@server1:~$ ip addr add 192.168.88.1/24 dev kvsbr0
# 把kvs网卡加入桥接
user1@server1:~$ brctl addif kvsbr0 kvs0
user1@server1:~$ brctl show kvsbr0
bridge name	bridge id		STP enabled	interfaces
kvsbr0		8000.de0bf784b5a9	no		kvs0

user1@server1:~$ ip addr show
...
8: kvs0: <BROADCAST,UP,LOWER_UP> mtu 1500 qdisc fq_codel master kvsbr0 state UNKNOWN group default qlen 1000
    link/ether de:0b:f7:84:b5:a9 brd ff:ff:ff:ff:ff:ff
    inet6 fe80::dc0b:f7ff:fe84:b5a9/64 scope link 
       valid_lft forever preferred_lft forever
9: kvsbr0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default qlen 1000
    link/ether de:0b:f7:84:b5:a9 brd ff:ff:ff:ff:ff:ff
    inet 192.168.88.1/24 scope global kvsbr0
       valid_lft forever preferred_lft forever
    inet6 fe80::4e0:28ff:fe79:4663/64 scope link 
       valid_lft forever preferred_lft forever
...
```

### 客户端
```shell script
user1@client1:~$ ip addr show
...
4: kvc0: <BROADCAST,UP,LOWER_UP> mtu 1500 qdisc fq_codel state UNKNOWN group default qlen 1000
    link/ether b6:ec:1d:b3:e6:90 brd ff:ff:ff:ff:ff:ff
    inet 192.168.88.2/24 brd 192.168.88.255 scope global kvc0
       valid_lft forever preferred_lft forever
    inet6 fe80::b4ec:1dff:feb3:e690/64 scope link 
       valid_lft forever preferred_lft forever
...

user1@client1:~$ ping 192.168.88.1
PING 192.168.88.1 (192.168.88.1) 56(84) bytes of data.
64 bytes from 192.168.88.1: icmp_seq=1 ttl=64 time=2.76 ms
64 bytes from 192.168.88.1: icmp_seq=2 ttl=64 time=0.881 ms
64 bytes from 192.168.88.1: icmp_seq=3 ttl=64 time=1.88 ms
64 bytes from 192.168.88.1: icmp_seq=4 ttl=64 time=1.75 ms
64 bytes from 192.168.88.1: icmp_seq=5 ttl=64 time=1.88 ms
^C
--- 192.168.88.1 ping statistics ---
5 packets transmitted, 5 received, 0% packet loss, time 4017ms
rtt min/avg/max/mdev = 0.881/1.834/2.764/0.599 ms

user1@client1:~$ cat /proc/net/arp | grep 192.168.88.1
192.168.88.1     0x1         0x2         de:0b:f7:84:b5:a9     *        kvc0
```

# F.A.Q
* 能否关闭服务器的IP分配功能？
    
    去除--assignable-ips参数即可，这样客户端将可以指定使用任何IP，tap模式下服务器也可使用DHCP为客户端分配IP

* 服务器通过非root权限运行，hook是通过非root身份调用的，导致hook无法调用root才能调用的接口

    如果hook是script，考虑通过sudo的NOPASSWD实现，如果hook是二进制可执行程序，还可通过capability或者SUID实现，因此请确保hook目录仅KCPVPN所运行的用户拥有访问权限！
    
* 如何提高吞吐量？

    程序默认使用的虚拟网卡MTU是1500，如果传输大量数据，程序需要频繁调用内核API，因此可尝试添加以下参数启用jumbo frame，把MTU设至9000：
    ```
  --vni-mtu 9000
  ```
    
    9000 / 1500 = 6，这意味着每个数据包可最多减少五次相关的系统调用
    
* 虚拟网卡MTU小于1500导致部分TCP服务无法正常使用
    
    请使用iptables修改SYN包的MSS值，取值一般为虚拟网卡的MTU - 40，例如--vni-mtu 1450，MSS取值1410：
    ```
  # -i与-o后面的“kvs”是虚拟网卡名称前缀，即--vni-name-prefix参数所指定的值
  iptables -t mangle -A FORWARD -i kvs+ -p tcp --syn -j TCPMSS --set-mss MSS值
  iptables -t mangle -A FORWARD -o kvs+ -p tcp --syn -j TCPMSS --set-mss MSS值
  ```
* Rapid FEC?

    通过--rapid-fec参数可以启用此模式，会增加少许带宽消耗，但可大幅度降低低pps情况下丢包对延迟的影响。效果如下图，左上未启用，右上已启用，下方裸线路：
![](https://cms.yuninter.net/wp-content/uploads/2019/11/mtr.png)
