package tunnel

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/satori/go.uuid"
)

// RegHandler 服务注册
func (tu *Tunnel) RegHandler() {
	// 获取数据结构
	tu.mux.HandleFunc("/getremoteserverlist", tu.getRemoteServerList)
	tu.mux.HandleFunc("/delRemoteServer", tu.delRemoteServer)
	tu.mux.HandleFunc("/addRemoteServer", tu.addRemoteServer)
	tu.mux.HandleFunc("/operateRemoteServer", tu.operateRemoteServer)
	tu.mux.HandleFunc("/login.json", tu.login)
}

func (tu *Tunnel) login(w http.ResponseWriter, r *http.Request) {
	param, err := getParams(r)
	if err != nil {
		w.Write(webResult(1, err.Error(), nil))
		return
	}
	if param.USER == tu.serverConfig.Admin.User && param.PASS == tu.serverConfig.Admin.Pwd {
		w.Write(webResult(0, ``, nil))
	} else {
		w.Write(webResult(1, `用户名或者密码不正确`, nil))
	}
}

// operateRemoteServer operate: stop/start
func (tu *Tunnel) operateRemoteServer(w http.ResponseWriter, r *http.Request) {
	param, err := getParams(r)
	if err != nil {
		w.Write(webResult(1, err.Error(), nil))
		return
	}
	if param.UUID == `` {
		w.Write(webResult(1, `非法请求`, nil))
		return
	}
	// 防止map并发读写
	tu.mutex.Lock()
	defer tu.mutex.Unlock()
	switch param.Operate {
	case `stop`:
		if !tu.remoteServer[param.UUID].Status {
			w.Write(webResult(1, `已经停止转发,请不要重复关闭...`, nil))
			return
		}
		if _, ok := tu.remoteServer[param.UUID]; ok {
			tu.remoteServer[param.UUID].quit <- true
		}
	case `start`:
		if tu.remoteServer[param.UUID].Status {
			w.Write(webResult(1, `该端口正在转发中...`, nil))
			return
		}
		if _, ok := tu.remoteServer[param.UUID]; ok {
			tu.remoteServer[param.UUID].Status = true
			go tu.runRemotePort(tu.remoteServer[param.UUID])
		}
	}
	// fmt.Println(`http://` + tu.remoteServer[param.UUID].LocalAddr)
	client := http.Client{
		Timeout: time.Second,
	}
	client.Get(`http://` + tu.remoteServer[param.UUID].LocalAddr)
	time.Sleep(2 * time.Second)
	var result = make([]interface{}, 0)
	for _, item := range tu.remoteServer {
		result = append(result, item)
	}
	w.Write(webResult(0, ``, result))
}

// getremoteServerList 获取remoteserver list
func (tu *Tunnel) getRemoteServerList(w http.ResponseWriter, r *http.Request) {
	// 防止map并发读写
	tu.mutex.RLock()
	defer tu.mutex.RUnlock()
	var result = make([]interface{}, 0)
	for _, item := range tu.remoteServer {
		result = append(result, item)
	}
	w.Write(webResult(0, ``, result))
}

// delRemoteList 删除不需要转发的端口
func (tu *Tunnel) delRemoteServer(w http.ResponseWriter, r *http.Request) {
	param, err := getParams(r)
	if err != nil {
		w.Write(webResult(1, err.Error(), nil))
		return
	}
	if param.UUID == `` {
		w.Write(webResult(1, `非法请求...`, nil))
		return
	}
	// 防止map并发读写
	tu.mutex.Lock()
	if _, ok := tu.remoteServer[param.UUID]; ok {
		if tu.remoteServer[param.UUID].Status {
			w.Write(webResult(1, `请先停止该端口的转发...`, nil))
			return
		}
		delete(tu.remoteServer, param.UUID)
	}
	tu.mutex.Unlock()
	w.Write(webResult(0, ``, map[string]interface{}{
		"code": "删除成功",
	}))
}

// addRemoteServerList 添加需要转发的端口
func (tu *Tunnel) addRemoteServer(w http.ResponseWriter, r *http.Request) {
	param, err := getParams(r)
	if err != nil {
		w.Write(webResult(1, err.Error(), nil))
		return
	}
	if param.UUID == `` {
		param.UUID = uuid.NewV4().String()
	}
	var status = true
	var message = ``
	clientConfig, err := buildSSHConfig(param.SSHUser, param.SSHPubkey, param.SSHPwd)
	if err != nil {
		status, message = false, err.Error()
	}
	tu.mutex.Lock()
	if server, ok := tu.remoteServer[param.UUID]; ok {
		if server.Status {
			w.Write(webResult(1, "请先停止该端口映射...", nil))
			tu.mutex.Unlock()
			return
		}
	}
	tu.remoteServer[param.UUID] = &RemoteServer{
		LocalAddr:    param.LocalAddr,
		RemoteAddr:   param.RemoteAddr,
		MiddleAddr:   param.MiddleAddr,
		sessions:     make(map[int32]*Session, 0),
		Status:       status,
		Message:      message,
		clientConfig: clientConfig,
		quit:         make(chan bool, 1),
		SSHpubkey:    param.SSHPubkey,
		SSHPwd:       param.SSHPwd,
		SSHUser:      param.SSHUser,
		UUID:         param.UUID,
	}
	tu.mutex.Unlock()
	if err := writeConfigFile(tu.filePath, tu.remoteServer[param.UUID]); err != nil {
		w.Write(webResult(1, err.Error(), nil))
		return
	}
	go tu.runRemotePort(tu.remoteServer[param.UUID])
	time.Sleep(time.Second)
	tu.mutex.Lock()
	var result = make([]interface{}, 0)
	for _, item := range tu.remoteServer {
		result = append(result, item)
	}
	tu.mutex.Unlock()
	w.Write(webResult(0, ``, result))
}

// 获取前端传递过来的参数
func getParams(r *http.Request) (param Params, err error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return param, err
	}
	if err = json.Unmarshal(body, &param); err != nil {
		return param, err
	}
	return param, nil
}

// Params 前端传过来的参数
type Params struct {
	Operate    string `json:"operate"` //参数(start|stop)
	UUID       string `json:"uuid"`    // 端口转发默认值
	LocalAddr  string `json:"localAddr"`
	MiddleAddr string `json:"middleAddr"`
	RemoteAddr string `json:"remoteAddr"`
	SSHUser    string `json:"sshuser"`
	SSHPwd     string `json:"sshpwd"`
	SSHPubkey  string `json:"sshpubkey"`
	PASS       string `json:"pass"`
	USER       string `json:"user"`
}

func webResult(result int, errorMsg string, data interface{}) []byte {
	web, _ := json.Marshal(WebResult{
		Result:   result,
		ErrorMsg: errorMsg,
		Data:     data,
	})
	return web
}

// WebResult ...
type WebResult struct {
	Result   int         `json:"result"`
	ErrorMsg string      `json:"errormsg"`
	Data     interface{} `json:"data"`
}

// 把新添加的数据写入配置文件
func writeConfigFile(filePath string, remoteServer *RemoteServer) (err error) {
	if filePath, err = filepath.Abs(filePath); err != nil {
		return err
	}
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	var serverconfig ServerConfig
	if err := json.Unmarshal(content, &serverconfig); err != nil {
		return err
	}
	lhost, lport, _ := net.SplitHostPort(remoteServer.LocalAddr)
	rhost, rport, _ := net.SplitHostPort(remoteServer.RemoteAddr)
	_, mport, _ := net.SplitHostPort(remoteServer.MiddleAddr)
	// net.SplitHostPort(remoteServer.LocalAddr)
	sshremote := SSHremote{
		LocalAddr:     lhost,
		LocalPort:     lport,
		RemoteAddr:    rhost,
		RemotePort:    rport,
		RemoteSSHPort: mport,
		SSHPwd:        remoteServer.SSHPwd,
		SSHUser:       remoteServer.SSHUser,
		SSHpubkey:     remoteServer.SSHpubkey,
		UUID:          remoteServer.UUID,
	}
	serverconfig.SSHremote = append(serverconfig.SSHremote, sshremote)
	result, err := json.Marshal(serverconfig)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filePath, result, os.ModeAppend)
}
