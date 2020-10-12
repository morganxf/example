package filepath

import "os"

// 注意Base("a/") == "a" 而不是""
func Base(path string) string {
	if path == "" {
		return "."
	}
	// 去除尾部多余的分隔符
	for len(path) > 0 && os.IsPathSeparator(path[len(path)-1]) {
		path = path[:len(path)-1]
	}
	if path == "" {
		return string(os.PathSeparator)
	}
	i := len(path)-1
	for i>=0 && !os.IsPathSeparator(path[i]) {
		i--
	}
	return path[i+1:]
}