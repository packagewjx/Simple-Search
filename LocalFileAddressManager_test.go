package simple_search

import (
	"bufio"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestReadFile(t *testing.T) {
	NewLocalFileAddressManager("dizhi.txt")
}

func TestReadAddresses(t *testing.T) {
	buf := make([]string, 0, 100)
	readAddresses("test.txt", 100, &buf)
	if len(buf) != 100 {
		t.Error("地址大小不对")
	}
	for _, add := range buf {
		fmt.Println(add)
	}
}

func TestSaveAddresses(t *testing.T) {
	buf := make([]string, 0, 200)
	file, _ := os.Open("dizhi.txt")
	reader := bufio.NewReader(file)
	for line, e := reader.ReadString('\n'); e == nil; line, e = reader.ReadString('\n') {
		buf = append(buf, line[:len(line)-1])
	}

	saveAddresses("out.txt", buf)
}

func TestLocalFileAddressManager(t *testing.T) {
	manager := NewLocalFileAddressManager("test.txt")

	consumer := func(index int) {
		_, givingChan, notifyGivingChan := manager.GetChannels()

		for true {
			notifyGivingChan <- true
			add := <-givingChan
			fmt.Println(index, "得到地址", add)
			time.Sleep(500 * time.Microsecond)
		}
	}

	go consumer(1)
	go consumer(2)
	go consumer(3)
	go consumer(4)
	go consumer(5)

	time.Sleep(5 * time.Second)
	go func() {
		fmt.Println("正在给予新的地址")
		savingChan, _, _ := manager.GetChannels()
		for i := 0; i < 100; i++ {
			savingChan <- "http://www.baidu.com"
		}
	}()

	time.Sleep(5 * time.Second)
	manager.Close()
}
