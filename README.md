## goSSHTunnel

> SSH的的Port Forward，中文可以称为端口转发，是SSH的一项非常重要的功能。它可以建立一条安全的SSH通道，并把任意的TCP连接放到这条通道中。
Port Forward SSH, Chinese can be referred to as port forwarding, is a very important feature of SSH. It can establish a secure SSH channel, and put any TCP connection into this channel.

## SSH端口映射 
`ssh -C -f -N -g -L 本地端口:目标IP:目标端口 用户名@目标IP`
`ssh -C -f -N -g -L 5678:10.16.93.204:9200 root@10.16.93.204`

## Example
有机器A[10.16.31.56]，B[10.16.93.204]。现想A通过ssh免密码登录到B

* 在A机下生成公钥/私钥对  `ssh-keygen -t rsa -P ''`

* 把A机下的id_rsa.pub复制到B机下`scp ~/.ssh/id_rsa.pub root@10.16.93.204:/root/`

* B机把从A机复制的id_rsa.pub添加到.ssh/authorzied_keys文件里`cat id_rsa.pub >> .ssh/authorized_keys`

* 把B机的authorized_keys文件夹加上600权限`chmod 600 .ssh/authorized_keys`

* 实现免密码登录验证`ssh root@10.16.93.204`


## sshtunnel


* 使用方法`go get -u github.com/farmerx/sshtunnel`

* 根据你的需要编译成不同平台下的程序
```
GO_ENABLED=0 GOOS=windows GOARCH=amd64  go build -o sshtunnel.exe github.com/sshtunnel
GO_ENABLED=0 GOOS=windows GOARCH=386  go build -o gosshtunnel.exe github.com/sshtunnel
GO_ENABLED=0 GOOS=linux GOARCH=amd64  go build -o sshtunnel github.com/sshtunnel
GO_ENABLED=0 GOOS=linux GOARCH=386  go build -o sshtunnel github.com/sshtunnel
```

* 获取目标目录器的ssh rsa公钥到本地，把路径填写到SSHTunnel.conf中， 并在里面配置你要映射端口的目标服务器用户名、ssh端口、要监听的端口，和你要映射到本地那个端口，如下：
```
{
    "Username": "root", 
    "PublicKeyPath": "/Users/farmerx/.ssh/id_rsa", 
    "ServerAddrString": "10.16.93.204:22", 
    "RemoteAddrString": "10.16.93.204:9200"
    "LocalAddrString": "127.0.0.1:3690", 
}

