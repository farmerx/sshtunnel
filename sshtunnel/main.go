package main

import (
	"encoding/json"
	_ "github.com/farmerx/sshtunnel/sshtunnel/statik"
	"github.com/farmerx/sshtunnel/sshtunnel/tunnel"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/farmerx/glog"
	"github.com/rakyll/statik/fs"
)

func main() {
	fmt.Println(`SSHTunnel v1.0.0`)
	runtime.GOMAXPROCS(runtime.NumCPU())
	// 解析flag参数
	parseFlags()
	// 解析配置文件
	if err := readServerConf(); err != nil {
		glog.Errorln(err)
		return
	}
	var mux = http.NewServeMux()
	statikFS, err := fs.New() // 静态文件编译成二进制
	if err != nil {
		glog.Errorln(err)
		return
	}
	http.FileServer(statikFS) // 加载http server file
	// rmux := mux.NewRouter()
	mux.Handle("/", http.StripPrefix("/", http.FileServer(statikFS)))
	mux.Handle("/login", http.StripPrefix("/login", http.FileServer(statikFS)))
	// 加载search package
	mux.HandleFunc("/ping", ping)

	// 初始化tunnel
	tunnel.NewTunnel(tunnel.InitOptions{
		Mux:          mux,
		FilePath:     config.FilePath,
		ServerConfig: config.ServerConfig,
	})
	// 初始化tunnel结束
	go func(addr string, rmux *http.ServeMux) {
		http.ListenAndServe(addr, rmux)
	}(config.ServerConfig.HTTPServer.Host+`:`+config.ServerConfig.HTTPServer.Port, mux)
	openPage()
	handleSignals()
}

// 定义一个server config的全局变量
var config *Config

// Config 配置文件
type Config struct {
	FilePath     string `json:"filepath"`
	ServerConfig tunnel.ServerConfig
}

// 接受退出信号
func handleSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
}

// ping pong
func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("PONG!\n"))
}

// readServerConf 解析配置文件
func readServerConf() (err error) {
	if config.FilePath, err = filepath.Abs(config.FilePath); err != nil {
		return err
	}
	content, err := ioutil.ReadFile(config.FilePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(content, &config.ServerConfig)
}

// parseFlags parse flags of program.
func parseFlags() {
	config = &Config{}
	flag.StringVar(&config.FilePath, "c", "./", "Configuration file address")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
}

func openPage() {
	url := fmt.Sprintf("http://%v", config.ServerConfig.HTTPServer.Host+":"+config.ServerConfig.HTTPServer.Port)
	fmt.Println("To view elasticHD console open", url, "in browser")
	var err error
	switch runtime.GOOS {
	case "linux":
		err = runCmd("xdg-open", url)
	case "darwin":
		err = runCmd("open", url)
	case "windows":
		r := strings.NewReplacer("&", "^&")
		err = runCmd("cmd", "/c", "start", r.Replace(url))
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		fmt.Println(err)
	}
}

var (
	Stderr io.Writer = os.Stderr // Stderr is the io.Writer to which executed commands write standard error.
	Stdout io.Writer = os.Stdout // Stdout is the io.Writer to which executed commands write standard output.
)

// runCmd run command opens a new browser window pointing to url.
func runCmd(prog string, args ...string) error {
	cmd := exec.Command(prog, args...)
	cmd.Stdout = Stdout
	cmd.Stderr = Stderr
	return cmd.Run()
}
