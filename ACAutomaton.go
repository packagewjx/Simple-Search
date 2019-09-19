package simple_search

type acNode struct {
	char   byte
	length int
	next   map[byte]*acNode
	isEnd  bool
	fail   *acNode
}

type ACFindResult struct {
	pos    int
	length int
}

// 多模式串匹配算法AC自动机
type ACAutomaton struct {
	root *acNode
}

// 构建AC自动机。包括构建前缀树等操作
func NewACAutomaton(words []string) *ACAutomaton {
	node := &acNode{
		char:   0,
		length: 0,
		next:   make(map[byte]*acNode),
		isEnd:  false,
		fail:   nil,
	}
	node.fail = node
	ac := &ACAutomaton{root: node}
	for _, word := range words {
		ac.addWord(word)
	}
	ac.setFail()
	return ac
}

// 使用广度遍历的方式设置fail指针
func (ac *ACAutomaton) setFail() {
	queue := make([]*acNode, 0, 16)
	queue = append(queue, ac.root)
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for char, node := range cur.next {
			if cur == ac.root {
				node.fail = ac.root
			} else {
				p := cur.fail
				for p != ac.root && p.next[char] == nil {
					p = p.fail
				}
				if p.next[char] == nil {
					node.fail = ac.root
				} else {
					node.fail = p.next[char]
				}
			}
			queue = append(queue, node)
		}
	}
}

// 加入新的单词，构建字典树
func (ac *ACAutomaton) addWord(word string) {
	if word == "" {
		return
	}
	cur := ac.root
	for i := 0; i < len(word)-1; i++ {
		if cur.next == nil {
			//若没创建，可以创建一个
			cur.next = make(map[byte]*acNode)
		}
		if cur.next[word[i]] == nil {
			cur.next[word[i]] = &acNode{
				char:   word[i],
				length: i + 1,
				next:   make(map[byte]*acNode),
				isEnd:  false,
				fail:   nil,
			}
		}
		cur = cur.next[word[i]]
	}
	if cur.next == nil {
		cur.next = make(map[byte]*acNode)
	}
	cur.next[word[len(word)-1]] = &acNode{
		char:   word[len(word)-1],
		length: len(word),
		next:   nil,
		isEnd:  true,
		fail:   nil,
	}
	ac.root.next[word[0]].fail = ac.root
}

func (ac *ACAutomaton) Add(word string) {
	ac.addWord(word)
	// 每加一次都要重新设置
	ac.setFail()
}

// 寻找模式串在s中出现的下标
func (ac *ACAutomaton) Find(s string) []*ACFindResult {
	res := make([]*ACFindResult, 0, 10)
	p := ac.root
	for i := 0; i < len(s); i++ {
		for p != ac.root && p.next[s[i]] == nil {
			p = p.fail
		}
		if p.next[s[i]] != nil {
			// 找到了匹配项
			p = p.next[s[i]]
			// 查看一系列fail指针是否是结尾
			q := p
			for q != ac.root {
				if q.isEnd {
					res = append(res, &ACFindResult{
						pos:    i - q.length + 1,
						length: q.length,
					})
				}
				q = q.fail
			}
		}
	}
	return res
}
