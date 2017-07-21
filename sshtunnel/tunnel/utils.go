package tunnel

import (
	"errors"
	"io/ioutil"
	"net"
	"os"

	"github.com/farmerx/glog"
	"golang.org/x/crypto/ssh"
)

func buildSSHConfig(user, pubkey, pwd string) (cfg *ssh.ClientConfig, err error) {
	if pubkey == `` && pwd == `` {
		return nil, errors.New(`公钥和密码全部为空`)
	}
	var signers []ssh.Signer
	// 优先使用公钥登录
	if len(pubkey) > 0 {
		if _, err := os.Stat(pubkey); os.IsNotExist(err) {
			return nil, err
		}
		pemBytes, err := ioutil.ReadFile(pubkey)
		if err != nil {
			return nil, err
		}
		if signer, err := ssh.ParsePrivateKey(pemBytes); err == nil {
			signers = append(signers, signer)
		}
	}
	// 不存在公钥使用密码登录
	var auths []ssh.AuthMethod
	if len(signers) < 1 {
		auths = []ssh.AuthMethod{ssh.Password(pwd)}
	} else {
		auths = []ssh.AuthMethod{ssh.PublicKeys(signers...)}
	}
	cfg = &ssh.ClientConfig{
		User: user,
		Auth: auths,
	}
	cfg.SetDefaults()
	return cfg, nil
}

func dialMiddleServer(remoteServer *RemoteServer) (sclient *ssh.Client, err error) {
	retry := 0
	for retry < 3 {
		glog.Infof("Connecting middle server %s...", remoteServer.MiddleAddr)
		sclient, err = ssh.Dial("tcp", remoteServer.MiddleAddr, remoteServer.clientConfig)
		if err == nil {
			break
		}
		glog.Errorf("Failed to connect middle server, err: %s, retry...", err.Error())
		retry++
	}
	return
}

func dialRemoteServer(sclient *ssh.Client, remoteServer *RemoteServer) (rconn net.Conn, err error) {
	retry := 0
	mh, _, _ := net.SplitHostPort(remoteServer.MiddleAddr)
	for retry < 3 {
		glog.Infof("Connecting remote server %s on %s...", remoteServer.RemoteAddr, mh)
		rconn, err = sclient.Dial("tcp", remoteServer.RemoteAddr)
		if err == nil {
			break
		}
		glog.Errorf("Failed to connect remote server %s, error: %s", remoteServer.RemoteAddr, err.Error())
		retry++
	}
	return
}
