package simple_search

// 拥有双向查找功能的map
type DualMap interface {
	// 若存在，则返回。若key对应的对象不存在，则返回空
	Get(key int) interface{}

	// 获取val对应的key。若val之前没有被保存，则会产生新的key，并返回。
	GetKey(val interface{}) int
}
