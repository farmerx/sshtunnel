package tunnel

import (
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/farmerx/glog"
)

// InitOptions ...
type InitOptions struct {
	Mux          *http.ServeMux
	ServerConfig ServerConfig
	FilePath     string
}

// Tunnel ...
type Tunnel struct {
	mux          *http.ServeMux
	serverConfig ServerConfig
	filePath     string
	remoteServer map[string]*RemoteServer
	mutex        sync.RWMutex
}

// NewTunnel ...
func NewTunnel(options InitOptions) *Tunnel {
	tu := new(Tunnel)
	tu.mux = options.Mux
	tu.serverConfig = options.ServerConfig
	tu.filePath = options.FilePath
	tu.remoteServer = make(map[string]*RemoteServer)
	// 初始化配置文件里面的配置
	tu.initialSSHClientConfig()
	return tu
}

// initialSSHClientConfig 初始化sshclientconfig
func (tu *Tunnel) initialSSHClientConfig() {
	for _, item := range tu.serverConfig.SSHremote {
		var status = true
		var message = ``
		clientConfig, err := buildSSHConfig(item.SSHUser, item.SSHpubkey, item.SSHPwd)
		if err != nil {
			status, message = false, err.Error()
		}
		tu.remoteServer[item.UUID] = &RemoteServer{
			LocalAddr:    item.LocalAddr + `:` + item.LocalPort,
			RemoteAddr:   item.RemoteAddr + `:` + item.RemotePort,
			MiddleAddr:   item.RemoteAddr + `:` + item.RemoteSSHPort,
			sessions:     make(map[int32]*Session, 0),
			Status:       status,
			Message:      message,
			clientConfig: clientConfig,
			quit:         make(chan bool, 1),
			SSHpubkey:    item.SSHpubkey,
			SSHPwd:       item.SSHPwd,
			SSHUser:      item.SSHUser,
			UUID:         item.UUID,
		}
	}
	tu.start()
	tu.RegHandler()
}

func (tu *Tunnel) start() {
	for _, item := range tu.remoteServer {
		if item.Status {
			go tu.runRemotePort(item)
		}
	}
}

// 端口转发 ,初始化好多结构体
func (tu *Tunnel) runRemotePort(remoteserver *RemoteServer) {
	addr, err := net.ResolveTCPAddr("tcp", remoteserver.LocalAddr)
	if err != nil {
		remoteserver.Status = false
		remoteserver.Message = err.Error()
		return
	}
	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		remoteserver.Status = false
		remoteserver.Message = err.Error()
		return
	}
	defer ln.Close()
	remoteserver.Message = fmt.Sprintf("Starting SSH Tunnel on %s...", remoteserver.LocalAddr)
	glog.Infof("Starting SSH Tunnel on %s...", remoteserver.LocalAddr)
	fmt.Println(fmt.Sprintf("Starting SSH Tunnel on %s...", remoteserver.LocalAddr))
	for {
		// accept tcp request
		conn, err := ln.AcceptTCP()
		if err != nil {
			glog.Infoln("accept:", err)
			continue
		}
		// accept quit single exit
		select {
		case <-remoteserver.quit:
			remoteserver.Status = false
			remoteserver.Message = fmt.Sprintf("Stop SSH Tunnel on %s...", remoteserver.LocalAddr)
			fmt.Println(fmt.Sprintf("Stop SSH Tunnel on %s...", remoteserver.LocalAddr))
			return
		default:
		}
		// data transport
		go transport(conn, remoteserver)
	}

}

func transport(lconn net.Conn, remoteServer *RemoteServer) {
	// Go中可以抛出一个panic的异常，然后在defer中通过recover捕获这个异常，然后正常处理
	// 在一个主进程，多个go程处理逻辑的结构中，这个很重要，如果不用recover捕获panic异常，会导致整个进程出错中断
	defer func() {
		if r := recover(); r != nil {
			glog.Errorln(strings.Join(traceInfo(), "\n\t"))
		}
		lconn.Close()
	}()

	start := time.Now()
	sclient, err := dialMiddleServer(remoteServer)
	if err != nil {
		glog.Errorf("Unable to connect middle server %s", remoteServer.MiddleAddr)
		return
	}

	rconn, err := dialRemoteServer(sclient, remoteServer)
	if err != nil {
		glog.Errorf("Unable to connect remote server %s", remoteServer.RemoteAddr)
		return
	}
	// defer rconn.Close()
	remoteServer.sessSeq++
	// t.sessSeq++
	sess := &Session{
		id:         remoteServer.sessSeq,
		tunnel:     remoteServer,
		localconn:  lconn,
		remoteconn: rconn,
		sshClient:  sclient,
		quit:       make(chan bool, 1),
	}

	putSession(sess, remoteServer)
	connectTime := time.Now().Sub(start)
	glog.Infof("[%d]Session created, %s, cost: %.2fs", sess.id, sess.directionInfo(), connectTime.Seconds())
	start = time.Now()
	sess.transferData() // block here
	transferTime := time.Now().Sub(start)

	glog.Infof("[%d]Session quit, connect-time: %.3fs transfer-time: %.3fs current-sessions-count: %d",
		sess.id, connectTime.Seconds(), transferTime.Seconds(), len(remoteServer.sessions))
}

func traceInfo() (trace []string) {
	for i := 1; ; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		trace = append(trace, fmt.Sprintf("%s:%d", file, line))
	}
	return
}
