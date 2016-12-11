package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"net"
	"sync"
	"time"

	"github.com/glog"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/crypto/ssh"
)

const (
	maxRetriesLocal  = 10 // 本地连接失败最大重试次数
	maxRetriesRemote = 5  // remote 最大重试次数
	maxRetriesServer = 5  // server 最大重试次数
	// SSHTUNNELCONFIGE config path

)

// ServerConf ...
type ServerConf struct {
	Username         string
	PublicKeyPath    string
	ServerAddrString string
	LocalAddrString  string
	RemoteAddrString string
}

// ReadServerConf read serverconf
func ReadServerConf() {
	selfDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	file, err := os.Open(filepath.Join(selfDir, configPath))
	if err != nil {
		glog.Fatalln(`配置文件不存在`)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&serverConf)
	if err != nil {
		glog.Fatalln(configPath, ` PARSE ERR:`, err)
	}
}

// TableRender 配置文件输出
func TableRender() {
	data := [][]string{
		[]string{"A", "Username", serverConf.Username},
		[]string{"B", "PublicKeyPath", serverConf.PublicKeyPath},
		[]string{"C", "ServerAddrString", serverConf.ServerAddrString},
		[]string{"D", "LocalAddrString", serverConf.LocalAddrString},
		[]string{"E", "RemoteAddrString", serverConf.RemoteAddrString},
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Num", "Name", "Value"})
	for _, v := range data {
		table.Append(v)
	}
	table.Render()
}

// ListenAndServer 创建监听本地端口
func ListenAndServer(localAddrString string, config *ssh.ClientConfig, serverAddrString, remoteAddrString string) error {
	localListener, err := net.Listen(`tcp`, localAddrString)
	if err != nil {
		return err
	}
	Server(localListener, config, serverAddrString, remoteAddrString)
	return nil
}

// Server ...
func Server(listen net.Listener, config *ssh.ClientConfig, serverAddrString, remoteAddrString string) {
	for {
		accept, err := listen.Accept()
		if err != nil {
			glog.Errorln("Accepting a client request failed:", err)
			continue
		}
		go forward(accept, config, serverAddrString, remoteAddrString)
	}
}

func forward(localConn net.Conn, config *ssh.ClientConfig, serverAddrString, remoteAddrString string) {
	defer localConn.Close()
	var sshClientConnection *ssh.Client
	currentRetriesServer := 0
	currentRetriesRemote := 0
	// 不断的循环重试,链接传承服务器
	for {
		sshClientConn, err := ssh.Dial(`tcp`, serverAddrString, config)
		if err != nil {
			currentRetriesServer++
			glog.Errorln("Was not able to connect with the SSH server ", serverAddrString, ":", err.Error())
			if currentRetriesServer < maxRetriesServer {
				glog.Infoln(`Retry...`)
				time.Sleep(1 * time.Second)
				continue
			}
			glog.Infoln(`No more retries for connecting the SSH server.`)
			return
		}
		glog.Infoln(`Connected to the SSH server ` + serverAddrString)
		sshClientConnection = sshClientConn
		defer sshClientConnection.Close()
		break
	}
	// ssh client 发送tcp request
	for {
		sshConn, err := sshClientConnection.Dial(`tcp`, remoteAddrString)
		if err != nil {
			currentRetriesRemote++
			glog.Errorln("Was not able to connect with the SSH server ", serverAddrString, ":", err.Error())
			if currentRetriesRemote < maxRetriesRemote {
				glog.Infoln(`Retry...`)
				time.Sleep(1 * time.Second)
				continue
			}
			glog.Errorln(`No more retries for connecting the SSH server.`)
			return
		}
		//端口已经连上
		glog.Infof("The remote end-point %s is connected.\n", remoteAddrString)
		defer sshConn.Close()
		var done *sync.WaitGroup
		done.Add(2)
		// io.copy local request write to remote request
		// io.copy remote response write to local response
		go transfer(localConn, sshConn, `Local => Remote`, done)
		go transfer(sshConn, localConn, `Remote => Local`, done)
		done.Wait()
		glog.Infoln(`Close now all connections.`)
		return
	}
}

// transfer ...
func transfer(fromReader io.Reader, toWriter io.Writer, name string, done *sync.WaitGroup) {
	defer done.Done()
	glog.Infoln("%s transfer started.", name)
	if _, err := io.Copy(toWriter, fromReader); err != nil {
		glog.Errorln(name, "transfer failed: \n", err.Error())
		return
	}
	glog.Infof("%s transfer closed.\n", name)
	return
}
