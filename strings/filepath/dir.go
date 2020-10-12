package filepath

import "os"

// Dir path的目录，尾缀不会包含多余的分隔符。
// 如果path以分隔符结束的话，Dir相当于只clean末尾多余的分隔符。Dir("a/")=="a"
func Dir(path string) string {
	i := len(path)-1
	// i 最小值为-1
	for i >= 0 && !os.IsPathSeparator(path[i]) {
		i--
	}
	// 此时i那么指向分隔符，要么为-1。为了处理-1的情况需要使用i+1作为范围终止
	// 即使使用i+1作为范围终止，最后一个字符可能是分隔符。不过Clean保证清理此种情况的尾缀分隔符
	return Clean(path[:i+1])
}