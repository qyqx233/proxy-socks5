package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	socks5 "github.com/armon/go-socks5"
)

func getIp() {
	fmt.Println("开始打印本机ip")
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				fmt.Println(ipnet.IP.String())
			}
		}
	}
}

func getPublicIp() {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("本机公网IP [%s]\n", string(ip))
}

type IpRule struct {
	AllowIp string
	regx    *regexp.Regexp
}

func NewIpControl(ip string) socks5.RuleSet {
	return &IpRule{
		AllowIp: ip,
		regx:    regexp.MustCompile(ip),
	}
}

type Syncer interface {
	Sync() (sl []string, err error)
}

type FileSyncer struct {
	f *os.File
}

func NewFileSyncer(path string) (fs *FileSyncer, err error) {
	var f *os.File
	f, err = os.Open(path)
	if err != nil {
		return
	}
	fs = &FileSyncer{
		f: f,
	}
	return
}

func (f *FileSyncer) Sync() (sl []string, err error) {
	r := bufio.NewReader(f.f)
	var data []byte
	for {
		data, _, err = r.ReadLine()
		if err != nil {
			if err == io.EOF {
				err = nil
				break
			}
		}
		sl = append(sl, string(bytes.Trim(data, " \t\r\n")))
	}
	return
}

func (c IpRule) Allow(ctx context.Context, req *socks5.Request) (context.Context, bool) {
	if c.regx.MatchString(req.RemoteAddr.IP.String()) {
		fmt.Println(req.DestAddr.FQDN)
		if len(whiteIpSet) == 0 {
			return ctx, true
		}
		if _, ok := whiteIpSet[req.DestAddr.FQDN]; ok {
			return ctx, true
		}
	}
	return ctx, false
}

var whiteIpSet map[string]struct{}

func init() {
	whiteIpSet = make(map[string]struct{})
}

func syncWhiteIp(file string) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	for _, ip := range strings.Split(string(data), "\n") {
		whiteIpSet[ip] = struct{}{}
	}
}

func syncWhiteIpLoop(file string) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		syncWhiteIp(file)
	}
}

func main() {
	var addr, port, allowIp, whiteIp string
	flag.StringVar(&addr, "addr", "localhost", "addr")
	flag.StringVar(&port, "port", "8082", "port")
	flag.StringVar(&allowIp, "aip", "8.8.8.8", "port")
	flag.StringVar(&whiteIp, "white", "white.ip", "white ip config")
	flag.Parse()
	getPublicIp()
	runtime.GOMAXPROCS(1)
	if _, err := os.Stat(whiteIp); err == nil {
		syncWhiteIp(whiteIp)
		go syncWhiteIpLoop(whiteIp)
	}
	conf := &socks5.Config{Rules: NewIpControl(allowIp)}
	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}
	// Create SOCKS5 proxy on localhost port 8000
	if err := server.ListenAndServe("tcp", fmt.Sprintf("%s:%s", addr, port)); err != nil {
		panic(err)
	}
}
