package simple_search

import (
	"bufio"
	"github.com/prometheus/common/log"
	"io"
	"os"
	"strconv"
	"sync"
)

const defaultSize = 1024

type newWordMessage struct {
	newKeyChan chan int
	word       string
}

// 基于数组的词语管理器
type ArrayWordManager struct {
	words       []string
	wordMap     sync.Map
	done        chan bool
	newWordChan chan *newWordMessage
	fileName    string
}

// 若存在，则返回词语。若key对应词语不存在，则返回空字符串
func (wm *ArrayWordManager) GetWord(key int) string {
	if key >= len(wm.words) {
		return ""
	}

	return wm.words[key]
}

func (wm *ArrayWordManager) GetKey(word string) int {
	key, ok := wm.wordMap.Load(word)
	if ok {
		return key.(int)
	} else {
		log.Info("接收到新词语", word)
		keyChan := make(chan int)
		wm.newWordChan <- &newWordMessage{
			newKeyChan: keyChan,
			word:       word,
		}
		// 再查询一次
		key := <-keyChan
		close(keyChan)
		log.Info("新词语", word, "的键为", key)
		return key
	}
}

func NewArrayWordManager(fileName string) *ArrayWordManager {
	// 读取原有的词语
	words, wordMap := readWords(fileName)

	wm := &ArrayWordManager{
		words:       words,
		wordMap:     wordMap,
		done:        make(chan bool),
		newWordChan: make(chan *newWordMessage),
		fileName:    fileName,
	}
	go appendWordRoutine(wm)
	return wm
}

func appendWordRoutine(wm *ArrayWordManager) {
	log.Info("词语管理器添加GoRoutine启动")
	for true {
		select {
		case <-wm.done:
			log.Info("词语管理器添加GoRoutine结束运行")
			return
		case msg := <-wm.newWordChan:
			val, ok := wm.wordMap.Load(msg.word)
			// 验证没有词语再添加
			if !ok {
				key := len(wm.words)
				log.Info("添加新词", msg.word, "，键值为", key)
				// 首先持久化到文件，然后再保存到内存。不过是让另一个goroutine执行，因此完成先后是不确定的。
				go appendWordToFile(key, msg.word, wm.fileName)
				wm.words = append(wm.words, msg.word)
				wm.wordMap.Store(msg.word, key)
				msg.newKeyChan <- key
			} else {
				msg.newKeyChan <- val.(int)
			}
		}
	}
}

func appendWordToFile(key int, word string, fileName string) {
	log.Info("正在保存词语", word, "与键值", key, "到文件", fileName)
	file, e := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if e != nil {
		panic(e)
	}
	keyString := strconv.Itoa(key)
	_, e = file.WriteString(keyString + "\t" + word + "\n")
	if e != nil {
		panic(e)
	}
	e = file.Close()
	if e != nil {
		panic(e)
	}
	log.Info("保存词语", word, "与键值", key, "到文件", fileName, "成功")
}

func readWords(fileName string) ([]string, sync.Map) {
	file, e := os.Open(fileName)
	if e != nil {
		log.Info("没有文件可以读取，返回中")
		return make([]string, 0, defaultSize), sync.Map{}
	}
	fin := bufio.NewReader(file)
	wordMap := sync.Map{}

	// 应付乱序保存的情况
	maxKey := -1
	handleLine := func(line string) {
		tab := 0
		for ; tab < len(line) && line[tab] != '\t'; tab++ {
		}
		key, _ := strconv.Atoi(line[:tab])
		word := line[tab+1:]
		wordMap.Store(word, key)
		if key > maxKey {
			maxKey = key
		}
	}

	// 读取
	line, e := fin.ReadString('\n')
	for e == nil {
		handleLine(line[:len(line)-1])
		line, e = fin.ReadString('\n')
	}
	if e == io.EOF && line != "" {
		handleLine(line)
	}

	// 填入数组
	capacity := defaultSize
	if maxKey+1 > capacity {
		// 多申请一些空间保存
		capacity = maxKey * 2
	}
	words := make([]string, maxKey+1, capacity)
	wordMap.Range(func(key, value interface{}) bool {
		word := key.(string)
		wordKey := value.(int)
		words[wordKey] = word
		log.Debug("读取到键值", wordKey, "词语", word)
		return true
	})
	log.Info("读取了", len(words), "个词语记录")

	return words, wordMap
}
