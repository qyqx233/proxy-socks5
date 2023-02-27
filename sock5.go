package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"runtime"
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
	enable  bool
}

func NewIpControl(ip string, enable bool) socks5.RuleSet {
	var regx *regexp.Regexp = nil
	if ip != "" {
		regx = regexp.MustCompile(ip)
	}
	return &IpRule{
		AllowIp: ip,
		regx:    regx,
		enable:  enable,
	}
}

func (c IpRule) Allow(ctx context.Context, req *socks5.Request) (context.Context, bool) {
	if c.regx != nil && !c.regx.MatchString(req.RemoteAddr.String()) {
		return ctx, false
	}
	if c.enable {
		if dt.nonEmpty {
			return ctx, dt.Contains(req.DestAddr.FQDN)
		}
		return ctx, false
	}
	return ctx, true
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

type DtSync struct {
	updatedAt time.Time
}

func (s *DtSync) syncWhiteIp(whiteIp string) {
	if stat, err := os.Stat(whiteIp); err == nil && stat.ModTime().After(s.updatedAt) {
		fmt.Println("更新ip白名单")
		syncer, err := NewFileSyncer(whiteIp)
		if err == nil {
			ips, err := syncer.Sync()
			if err == nil {
				for _, ip := range ips {
					dt.nonEmpty = true
					dt.tree.AddDomainSuffix(ip, "")
				}
				s.updatedAt = stat.ModTime()
			}
		}
	}
}

var dtSync = new(DtSync)

func syncWhiteIpLoop(file string) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		dtSync.syncWhiteIp(file)
	}
}

type emptyWriter struct {
}

func (w emptyWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func main() {
	var addr, port, allowIp, whiteIp string
	var verbose, dif bool
	flag.StringVar(&addr, "addr", "localhost", "addr")
	flag.StringVar(&port, "port", "8082", "port")
	flag.StringVar(&allowIp, "aip", "", "port")
	flag.StringVar(&whiteIp, "white", "white.ip", "white ip config")
	flag.BoolVar(&verbose, "v", false, "verbose")
	flag.BoolVar(&dif, "dif", true, "dest ip filter")
	flag.Parse()

	fmt.Printf("dif=[%v], verbose=[%v]\n", dif, verbose)
	getPublicIp()
	runtime.GOMAXPROCS(1)

	dtSync.syncWhiteIp(whiteIp)
	go syncWhiteIpLoop(whiteIp)
	conf := &socks5.Config{
		Rules: NewIpControl(allowIp, dif),
		// Logger: log.New(emptyWriter{}, "", log.LstdFlags),
	}
	if !verbose {
		fmt.Println("禁止日志输出")
		conf.Logger = log.New(emptyWriter{}, "", log.LstdFlags)
	}
	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}
	// Create SOCKS5 proxy on localhost port 8000
	if err := server.ListenAndServe("tcp", fmt.Sprintf("%s:%s", addr, port)); err != nil {
		panic(err)
	}
}
