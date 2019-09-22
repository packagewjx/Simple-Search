package simple_search

import (
	"bufio"
	"os"
	"testing"
	"time"
)

func TestArrayWordManager(t *testing.T) {
	manager := NewArrayWordManager("wordtest.txt", func(val interface{}) string {
		return val.(string)
	}, func(valString string) interface{} {
		return valString
	})

	words := make([]string, 0, 3000)
	file, _ := os.Open("words.txt")
	fin := bufio.NewReader(file)
	line, e := fin.ReadString('\n')
	for e == nil {
		words = append(words, line[:len(line)-1])
		line, e = fin.ReadString('\n')
	}
	if line != "" {
		words = append(words, line)
	}

	wordMap := make(map[string]int)
	for i := 0; i < len(words); i++ {
		key := manager.GetKey(words[i])
		wordMap[words[i]] = key
	}

	time.Sleep(5 * time.Second)

	for word, key := range wordMap {
		k := manager.GetKey(word)
		if k != key {
			t.Error(word, "的键值应该是", key, "而不是", k)
		}
	}

	for i := 0; i < len(manager.values); i++ {
		if manager.values[i] == "" {
			t.Error(i, "有空的部分")
		}
	}
}

// 测试读取
func TestNewArrayWordManager(t *testing.T) {
	manager := NewArrayWordManager("wordtest.txt", func(val interface{}) string {
		return val.(string)
	}, func(valString string) interface{} {
		return valString
	})
	for i := 0; i < len(manager.values); i++ {
		if manager.values[i] == "" {
			t.Error(i, "有空的部分")
		}
		key := manager.GetKey(manager.values[i])
		if key != i {
			t.Error("键值应该是", i, "而不是", key)
		}
	}
}
