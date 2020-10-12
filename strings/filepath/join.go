package filepath

import (
	"os"
	"strings"
)

func Join(elem ...string) string {
	for i, e := range elem {
		// 找到第一个非空的string再join
		// 针对绝对路径没有影响，因为Clean会消除掉重复的PathSeparator
		// 针对相对路径会有影响，如果第一个元素为空，那么join之后就会变为绝对路径
		if e != "" {
			return Clean(strings.Join(elem[i:], string(os.PathSeparator)))
		}
	}
	return ""
}
