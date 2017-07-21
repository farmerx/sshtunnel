// Listen on local port 9000.
// Upon attempted read from local port 9000: (listener.Accept()),
// Accept connection and return a local io.Reader and io.Writer and,
// Connect to remote server and,
// Connect to remote port 9999 returning a io.Reader and io.Writer.
// Continually copy local io.Reader bytes to remote io.Writer,
// Continually copy remote io.Reader bytes to local io.Writer.

package tunnel

import (
	"fmt"
	"io"
	"time"

	"github.com/farmerx/glog"
	"golang.org/x/crypto/ssh"
)

func putSession(sess *Session, remoteServer *RemoteServer) {
	remoteServer.mutex.Lock()
	remoteServer.sessions[sess.id] = sess
	remoteServer.mutex.Unlock()
}

func (s *Session) close() {
	s.remoteconn.Close()
	s.localconn.Close()
	s.sshClient.Close()
	removeSession(s)
}

func removeSession(sess *Session) {
	sess.tunnel.mutex.Lock()
	delete(sess.tunnel.sessions, sess.id)
	sess.tunnel.mutex.Unlock()
}

func (s *Session) directionInfo() string {
	return fmt.Sprintf("%s <==> %s <==> %s", s.tunnel.LocalAddr, s.tunnel.MiddleAddr, s.tunnel.RemoteAddr)
}
func (s *Session) transferData() {
	go func() {
		_, err := io.Copy(s.remoteconn, s.localconn)
		if err != nil {
			glog.Infoln(err.Error())
		}
		s.quit <- true
	}()
	go func() {
		_, err := io.Copy(s.localconn, s.remoteconn)
		if err != nil {
			glog.Infoln(err.Error())
		}
		s.quit <- true
	}()

	go func() {
		s.heartbeat()
	}()

	select {
	case <-s.quit:
		s.close()
	}
}

// SSHHeartbeatMsg ...
type SSHHeartbeatMsg struct {
	// See RFC 4253, section 12
	SSHType string `sshtype:"2"`
}

// heartbeat 心跳重连机制
func (s *Session) heartbeat() {
	// Send heartbeat packet every 10 seconds for SSH connection(ClientAliveInterval)
	payload := SSHHeartbeatMsg{SSHType: "2"}
	packet := ssh.Marshal(&payload)
	ticker := time.NewTicker(time.Duration(10) * time.Second)
	for {
		select {
		case <-ticker.C:
			// SendRequest sends a global request, and returns the
			// reply. If wantReply is true, it returns the response status
			// and payload. See also RFC4254, section 4.
			s.sshClient.Conn.SendRequest("hello", true, packet)
		case <-s.quit:
			glog.Infoln("quit heartbeat thread")
			return
		}
	}
}
