package main

import "testing"

func Test_white(t *testing.T) {
	syncWhiteIp("white.ip")
	t.Log(dt.Contains("www.baidu.com"))
	// t.Log("Test")
}
