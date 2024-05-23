package packet

// Overhead 表示一个对象对packet读写时的开销，
// Attach 表示读写时附加到包头的最大数据大小
// Append 表示读写时追加到包尾的最大数据大小
type Overhead interface {
	Overhead() (attach, append int)
}

func Inherit(child Overhead, attach, append int) (int, int) {
	h, t := child.Overhead()
	return h + attach, t + append
}
