package main

// domain_suffix_trie "github.com/golang-infrastructure/go-domain-suffix-trie"

type domainTester struct {
	nonEmpty bool
	tree     *DomainSuffixTrieNode[string]
}

func NewDomainTester() *domainTester {
	tree := NewDomainSuffixTrie[string]()
	return &domainTester{
		nonEmpty: false,
		tree:     tree,
	}
}

func (t *domainTester) Add(domain string) {
	t.nonEmpty = true
	t.tree.AddDomainSuffix(domain, "")
}

func (t *domainTester) Contains(domain string) bool {
	node := t.tree.FindMatchDomainSuffixNode(domain)
	return node != nil && node.isLast
}

// func (t *domainTester) Get(domain string) string {
// 	return t.tree.Get(domain)
// }

// var domainTester = NewDomainTester()

var dt = NewDomainTester()
