package main

import "testing"

func Test(t *testing.T) {
	tree := NewTrie()
	tree.Insert("abc")
	tree.Insert("ade")
	t.Log(tree.Find("ade"))
	t.Log(tree.Find("/"))
}

func Test_domain(t *testing.T) {
	tester := NewDomainTester()
	tester.Add("openai.com")
	t.Log(tester.Contains("chat.openai.com"))
	t.Log(tester.Contains("a.cn"))
	t.Log(tester.Contains("openai.com"))
	t.Log(tester.Contains("chat.qpenai.com"))
	// t.Log(tester.tree.FindMatchDomainSuffixNode("penai.com"))
}
