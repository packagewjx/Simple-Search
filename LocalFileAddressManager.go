package simple_search

import (
	"bufio"
	"github.com/prometheus/common/log"
	"io"
	"os"
)

const filterSize = 1 << 20
const bufferSize = 1 << 10

// 保存到文件中的阈值，只要超过这个阈值，就要保存这部分地址到文件中，以免丢失
const saveThreshold = 1 << 8

type LocalFileAddressManager struct {
	fileName         string
	filter           *BloomFilter
	addressBuffer    []string
	done             chan bool
	savingChan       chan string
	givingChan       chan string
	notifyGivingChan chan bool
}

func (am *LocalFileAddressManager) GetChannels() (savingChan chan<- string, givingChan <-chan string, notifyGivingChan chan<- bool) {
	return am.savingChan, am.givingChan, am.notifyGivingChan
}

func (am *LocalFileAddressManager) Close() {
	am.done <- true
	// 等待清理完成的信号
	<-am.done
}

func NewLocalFileAddressManager(fileName string) *LocalFileAddressManager {
	// 判断文件是否存在
	file, e := os.Open(fileName)
	if e != nil {
		panic(e)
	}

	info, e := file.Stat()
	if e != nil {
		panic(e)
	}
	length := info.Size()
	log.Info("地址文件", fileName, "长度", length)

	filter := NewBloomFilter(filterSize, DefaultHashFunc())
	var addBuf []string
	if length > 0 {
		log.Info("从旧文件中读取地址并恢复布隆过滤器")
		// 重建布隆过滤器
		reader := bufio.NewReader(file)
		line, e := reader.ReadString('\n')
		allAddresses := make([]string, 0, bufferSize)
		for e == nil {
			add := line[:len(line)-1]
			allAddresses = append(allAddresses, add)
			filter.Add([]byte(add))
			line, e = reader.ReadString('\n')
		}
		// 加上最后一个
		if e == io.EOF {
			if line != "" {
				add := line[:len(line)-1]
				allAddresses = append(allAddresses, add)
				filter.Add([]byte(add))
			}
		} else if e != nil {
			panic(e)
		}

		log.Info("成功读取", len(allAddresses), "条记录")
		addBufSize := saveThreshold >> 1
		var writeBuf []string
		if addBufSize > len(allAddresses) {
			addBuf = allAddresses
			// 截断
			writeBuf = []string{}
			log.Info("文件", fileName, "地址将清空")
		} else {
			// 保存后面多余的部分回去
			addBuf = allAddresses[:addBufSize]
			writeBuf = allAddresses[addBufSize:]
			log.Info("保存多出", len(writeBuf), "条地址到文件中")
		}

		// 保存多余的地址
		fout, e := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC, 0666)
		if e != nil {
			panic(e)
		}
		delim := []byte{'\n'}
		for _, add := range writeBuf {
			_, e = fout.WriteString(add)
			if e != nil {
				panic(e)
			}
			_, e = fout.Write(delim)
			if e != nil {
				panic(e)
			}
		}
		e = fout.Close()
		if e != nil {
			panic(e)
		}
	} else {
		addBuf = make([]string, 0, bufferSize)
	}
	e = file.Close()
	if e != nil {
		panic(e)
	}

	log.Info("地址管理器初始化完成")
	am := &LocalFileAddressManager{
		fileName:         fileName,
		filter:           filter,
		addressBuffer:    addBuf,
		done:             make(chan bool),
		savingChan:       make(chan string),
		givingChan:       make(chan string),
		notifyGivingChan: make(chan bool),
	}
	// 启动管理routine
	go managerRoutine(am)

	return am
}

func managerRoutine(am *LocalFileAddressManager) {
	waitingRoutines := 0
	log.Info("地址管理器主线程启动")
	for true {
		select {
		case <-am.done:
			// 完成退出
			log.Info("地址管理器收到退出命令")
			if len(am.addressBuffer) > 0 {
				log.Info("保存剩余的地址到文件中")
				saveAddresses(am.fileName, am.addressBuffer)
				log.Info("保存完成")
			}
			log.Info("地址管理器主线程退出")
			// 通知已完成清理
			am.done <- true
			return
		case add := <-am.savingChan:
			log.Info("获取到新的地址", add)
			am.addressBuffer = append(am.addressBuffer, add)
			if waitingRoutines > 0 {
				log.Info("有正在等待的Routines", waitingRoutines)
				add := am.addressBuffer[len(am.addressBuffer)-1]
				log.Debug("发送地址", add)
				am.addressBuffer = am.addressBuffer[:len(am.addressBuffer)-1]
				am.givingChan <- add
				waitingRoutines--
			} else if len(am.addressBuffer) > saveThreshold {
				// 保存
				log.Info("地址数量", len(am.addressBuffer), "超出限制", saveThreshold, "，保存多余的地址中")
				saveBuf := am.addressBuffer[:saveThreshold]
				am.addressBuffer = am.addressBuffer[saveThreshold:]
				saveAddresses(am.fileName, saveBuf)
				log.Info("地址保存完成")
			}
		case <-am.notifyGivingChan:
			log.Info("接收到地址请求")
			if len(am.addressBuffer) == 0 && waitingRoutines == 0 {
				// 在没有等待的时候，尝试读取新的地址
				// 若已经有等待的进程，说明是肯定是没有新的地址了
				log.Info("地址不足，读取新的地址中")
				readAddresses(am.fileName, saveThreshold>>1, &am.addressBuffer)
				log.Info("读取了", len(am.addressBuffer), "条地址")
			}
			if len(am.addressBuffer) > 0 {
				add := am.addressBuffer[len(am.addressBuffer)-1]
				log.Debug("发送地址", add)
				am.addressBuffer = am.addressBuffer[:len(am.addressBuffer)-1]
				am.givingChan <- add
			} else {
				// 彻底没有地址了，因此等待数加1，在接收到地址的时候再发送
				log.Info("无地址，goRoutine进入等待")
				waitingRoutines++
			}
		}
	}
}

func saveAddresses(fileName string, addresses []string) {
	var file *os.File
	file, e := os.OpenFile(fileName, os.O_APPEND, 0666)
	if e != nil {
		file, e = os.Create(fileName)
		if e != nil {
			panic(e)
		}
	}

	for _, add := range addresses {
		_, e = file.WriteString(add)
		if e != nil {
			panic(e)
		}
		_, e = file.Write([]byte{'\n'})
		if e != nil {
			panic(e)
		}
	}

	e = file.Close()
	if e != nil {
		panic(e)
	}
}

func readAddresses(fileName string, number int, buffer *[]string) {
	// 读取文件
	file, e := os.Open(fileName)
	if e != nil {
		panic(e)
	}
	fin := bufio.NewReader(file)
	for i := 0; i < number; i++ {
		line, e := fin.ReadString('\n')
		if e == io.EOF {
			// 没有读取足够数量就到文件尾
			if line != "" {
				*buffer = append(*buffer, line)
			}
			e = file.Close()
			if e != nil {
				panic(e)
			}
			// 删除剩余的
			fout, e := os.OpenFile(fileName, os.O_TRUNC|os.O_WRONLY, 0666)
			if e != nil {
				panic(e)
			}
			e = fout.Close()
			if e != nil {
				panic(e)
			}
			return
		} else if e != nil {
			panic(e)
		}
		*buffer = append(*buffer, line[:len(line)-1])
	}

	// 将剩余继续读取
	buf := make([]string, 0, number)
	line, e := fin.ReadString('\n')
	for e == nil {
		buf = append(buf, line)
		line, e = fin.ReadString('\n')
	}

	e = file.Close()
	fin = nil
	if e != nil {
		panic(e)
	}

	// 写回文件中
	file, e = os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC, 0666)
	if e != nil {
		panic(e)
	}
	for _, add := range buf {
		_, e = file.WriteString(add)
		if e != nil {
			panic(e)
		}
	}
	e = file.Close()
	if e != nil {
		panic(e)
	}
}
