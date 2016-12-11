package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/glog"
	"golang.org/x/crypto/ssh"
)

var (
	serverConf = &ServerConf{}
	configPath string
)

func init() {
	flag.Set("logtostderr", "1")
	flag.StringVar(&configPath, "f", `.`+string(os.PathSeparator)+`SSHTunnel.conf`, "user -f <config filepath>")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
}

func main() {
	// 打印出当前的版本信息:
	fmt.Println(`SSHTunnel v1.1.0`)
	runtime.GOMAXPROCS(runtime.NumCPU())
	ReadServerConf() //读取配置文件
	TableRender()    //打印配置文件
	privateKey, err := ioutil.ReadFile(serverConf.PublicKeyPath)
	if err != nil {
		glog.Warningln("id_rsa file not found :", err)
		os.Exit(0)
	}
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		glog.Warningln("The privateKey format is not correct :", err)
		os.Exit(0)
	}
	config := &ssh.ClientConfig{
		User: serverConf.Username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}
	// 创建一个本地连接:
	if err := ListenAndServer(serverConf.LocalAddrString, config, serverConf.ServerAddrString, serverConf.RemoteAddrString); err != nil {
		glog.Warningln(err)
		os.Exit(0)
	}
	// 捕获ctrl-c,平滑退出
	chExit := make(chan os.Signal, 1)
	signal.Notify(chExit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	select {
	case <-chExit:
	}
}
