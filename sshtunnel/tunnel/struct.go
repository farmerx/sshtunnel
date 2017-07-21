package tunnel

import (
	"net"
	"sync"

	"golang.org/x/crypto/ssh"
)

// ServerConfig 服务端配置文件
type ServerConfig struct {
	Admin struct {
		Pwd  string `json:"pwd"`
		User string `json:"user"`
	} `json:"admin"`
	HTTPServer struct {
		Host string `json:"host"`
		Port string `json:"port"`
	} `json:"httpServer"`
	SSHremote []SSHremote `json:"sshremote"`
}

// SSHremote ...
type SSHremote struct {
	LocalAddr     string `json:"localAddr"`
	LocalPort     string `json:"localPort"`
	RemoteAddr    string `json:"remoteAddr"`
	RemotePort    string `json:"remotePort"`
	RemoteSSHPort string `json:"remoteSSHPort"`
	SSHPwd        string `json:"sshPwd"`
	SSHUser       string `json:"sshUser"`
	SSHpubkey     string `json:"sshpubkey"`
	UUID          string `json:"uuid"`
}

// RemoteServer ...
type RemoteServer struct {
	UUID         string            `json:"uuid"`
	LocalAddr    string            `json:"localaddr"`  // 本地host
	RemoteAddr   string            `json:"remoteaddr"` // 远程host
	MiddleAddr   string            `json:"middleaddr"` // ssh远程服务器端口和ip
	clientConfig *ssh.ClientConfig // clientconfig
	Message      string            `json:"message"` // tunnel错误信息
	Status       bool              `json:"ststus"`  // tunnel当前状态
	SSHPwd       string            `json:"sshPwd"`
	SSHUser      string            `json:"sshUser"`
	SSHpubkey    string            `json:"sshpubkey"`
	sessSeq      int32
	sessions     map[int32]*Session
	mutex        sync.RWMutex
	quit         chan bool
}

// Session ...
type Session struct {
	id         int32
	tunnel     *RemoteServer
	localconn  net.Conn
	remoteconn net.Conn
	sshClient  *ssh.Client
	quit       chan bool
}
