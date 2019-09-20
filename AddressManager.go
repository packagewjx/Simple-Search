package simple_search

type AddressManager interface {
	// 获取与AddressManager沟通的通道。
	// savingChan保存新的地址。givingChan获取新的地址，但是需要首先发送信号给notifyGivingChan，才能获取
	GetChannels() (savingChan chan<- string, givingChan <-chan string, notifyGivingChan chan<- bool)

	// 保存仍未爬取的地址，进行资源回收的工作
	Close()
}
