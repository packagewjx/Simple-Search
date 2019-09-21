package simple_search

type WordManager interface {
	// 若存在，则返回词语。若key对应词语不存在，则返回空字符串
	GetWord(key int) string

	// 获取word对应的key。若word之前没有被保存，则会产生新的key，并返回。
	GetKey(word string) int
}
