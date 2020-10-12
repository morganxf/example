package filepath

import (
	"os"
	"sort"
	"strings"
)

// Glob 根据输入的pattern匹配所有的文件。如果没有匹配到，返回nil
//
// Glob 递归match
// 1. 假定第n级dirs已经匹配，那么就需要遍历dirs调用glob匹配file，得到第n+1 dirs
// 2. 递归终止条件，dir已经匹配展开
func Glob(pattern string) (matches []string, err error) {
	if !hasMeta(pattern) {
		if _, err = os.Lstat(pattern); err != nil {
			return nil, nil
		}
		return []string{pattern}, nil
	}

	// path中包含magic chars
	dir, file := Split(pattern)
	dir = cleanGlobPath(dir)

	// 递归终止条件
	// dir不包含魔法字符，处于已展开匹配状态
	if !hasMeta(dir) {
		return glob(dir, file, nil)
	}

	var m []string
	// 递归调用Glob
	// 由于dir包含魔法字符，需要递归处理dir，知道不包含魔法字符
	m, err = Glob(dir)
	if err != nil {
		return
	}
	// 递归后处理
	for _, d := range m {
		// 循环更新matches
		matches, err = glob(d, file, matches)
		if err != nil {
			return
		}
	}
	return
}

func cleanGlobPath(path string) string {
	switch path {
	case "":
		return "."
	case string(os.PathSeparator):
		return path
	default:
		// 去掉末尾的'/'。由Split知道，返回的dir末尾包含'/'
		return path[:len(path)-1]
	}
}

// glob dir已经匹配展开的情况下，寻找dir下匹配pattern的文件，并join增加到matches列表中.
// 如果存在问题，matches不变、返回。
func glob(dir, pattern string, matches []string) ([]string, error) {
	fi, err := os.Stat(dir)
	if err != nil {
		// matches不变、返回
		return matches, nil
	}
	if !fi.IsDir() {
		return matches, nil
	}
	d, err := os.Open(dir)
	if err != nil {
		return matches, nil
	}
	defer d.Close()
	names, _ := d.Readdirnames(-1)
	sort.Strings(names)
	for _, n := range names {
		matched, err := Match(pattern, n)
		if err != nil {
			return matches, err
		}
		if matched {
			matches = append(matches, Join(dir, n))
		}
	}
	return matches, nil
}

func hasMeta(path string) bool {
	magicChars := "*?["
	return strings.ContainsAny(path, magicChars)
}
