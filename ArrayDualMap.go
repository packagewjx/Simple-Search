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

type newValMessage struct {
	newKeyChan chan int
	val        interface{}
}

// 基于数组的词语管理器
type ArrayDualMap struct {
	values       []interface{}
	valueMap     sync.Map
	done         chan bool
	newValChan   chan *newValMessage
	fileName     string
	serializer   func(val interface{}) string
	deserializer func(valString string) interface{}
}

// 若存在，则返回词语。若key对应词语不存在，则返回空字符串
func (dMap *ArrayDualMap) Get(key int) interface{} {
	if key >= len(dMap.values) || key < 0 {
		return nil
	}

	return dMap.values[key]
}

func (dMap *ArrayDualMap) GetKey(val interface{}) int {
	key, ok := dMap.valueMap.Load(val)
	if ok {
		return key.(int)
	} else {
		log.Info("接收到新值", val)
		keyChan := make(chan int)
		dMap.newValChan <- &newValMessage{
			newKeyChan: keyChan,
			val:        val,
		}
		// 再查询一次
		key := <-keyChan
		close(keyChan)
		log.Info("新值", val, "的键为", key)
		return key
	}
}

func NewArrayWordManager(fileName string,
	serializer func(val interface{}) string,
	deserializer func(valString string) interface{}) *ArrayDualMap {
	dMap := &ArrayDualMap{
		done:         make(chan bool),
		newValChan:   make(chan *newValMessage),
		fileName:     fileName,
		serializer:   serializer,
		deserializer: deserializer,
	}
	dMap.readValues()
	go dMap.appendWordRoutine()
	return dMap
}

func (dMap *ArrayDualMap) appendWordRoutine() {
	log.Info("双向表添加GoRoutine启动")
	for true {
		select {
		case <-dMap.done:
			log.Info("双向表添加GoRoutine结束运行")
			return
		case msg := <-dMap.newValChan:
			val, ok := dMap.valueMap.Load(msg.val)
			// 验证没有词语再添加
			if !ok {
				key := len(dMap.values)
				valString := dMap.serializer(msg.val)
				log.Info("添加", msg.val, "，键值为", key)
				// 首先持久化到文件，然后再保存到内存。不过是让另一个goroutine执行，因此完成先后是不确定的。
				go dMap.saveRoutine(key, valString)
				dMap.values = append(dMap.values, msg.val)
				dMap.valueMap.Store(msg.val, key)
				msg.newKeyChan <- key
			} else {
				msg.newKeyChan <- val.(int)
			}
		}
	}
}

func (dMap *ArrayDualMap) saveRoutine(key int, valString string) {
	log.Info("正在保存", valString, "与键值", key, "到文件", dMap.fileName)
	file, e := os.OpenFile(dMap.fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if e != nil {
		panic(e)
	}
	keyString := strconv.Itoa(key)
	_, e = file.WriteString(keyString + "\t" + valString + "\n")
	if e != nil {
		panic(e)
	}
	e = file.Close()
	if e != nil {
		panic(e)
	}
	log.Info("保存", valString, "与键值", key, "到文件", dMap.fileName, "成功")
}

func (dMap *ArrayDualMap) readValues() {
	file, e := os.Open(dMap.fileName)
	dMap.valueMap = sync.Map{}
	if e != nil {
		dMap.values = make([]interface{}, 0, defaultSize)
		log.Info("没有文件可以读取，返回中")
		return
	}
	fin := bufio.NewReader(file)

	// 应付乱序保存的情况
	maxKey := -1
	handleLine := func(line string) {
		tab := 0
		for ; tab < len(line) && line[tab] != '\t'; tab++ {
		}
		key, _ := strconv.Atoi(line[:tab])
		valString := line[tab+1:]
		val := dMap.deserializer(valString)
		dMap.valueMap.Store(val, key)
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
	dMap.values = make([]interface{}, maxKey+1, capacity)
	dMap.valueMap.Range(func(val, key interface{}) bool {
		wordKey := key.(int)
		dMap.values[wordKey] = val
		log.Debug("读取到键值", wordKey, "与值", dMap.serializer(val))
		return true
	})
	log.Info("读取了", len(dMap.values), "个记录")
	return
}
