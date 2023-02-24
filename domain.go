package main

import (
	domain_suffix_trie "github.com/golang-infrastructure/go-domain-suffix-trie"
)

type domainTester struct {
	tree *domain_suffix_trie.DomainSuffixTrieNode[string]
}

func NewDomainTester() *domainTester {
	tree := domain_suffix_trie.NewDomainSuffixTrie[string]()
	return &domainTester{
		tree: tree,
	}
}

func (t *domainTester) Add(domain string) {
	t.tree.AddDomainSuffix(domain, "")
}

func (t *domainTester) Contains(domain string) bool {
	return t.tree.FindMatchDomainSuffixNode(domain) != nil
}

// func (t *domainTester) Get(domain string) string {
// 	return t.tree.Get(domain)
// }
