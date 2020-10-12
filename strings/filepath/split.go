package filepath

import (
	"os"
)

// Split从右向左迭代寻找filepath分隔符'/', 而非strings.Split(path, "/")
// 如果存在dir即不为空，则返回的dir以'/'符号结尾
//
// 为什么用"i+1": 因为当path[i]=='/'退出for循环时，此时file应该从i+1位置开始
// 为什么用"i>="作为for条件: 因为以i+1为分隔，所以i的最小取值应该为-1
func Split(path string) (dir, file string) {
	i := len(path) - 1
	for i >= 0 && !os.IsPathSeparator(path[i]) {
		i--
	}
	return path[:i+1], path[i+1:]
}
