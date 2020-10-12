package filepath

import (
	"errors"
	"os"
	"syscall"
)

var (
	ErrTooManyLinks = errors.New("EvalSymlinks: too many links")
)

func EvalSymlinks(path string) (string, error) {
	return walkSymlinks(path)
}

func walkSymlinks(path string) (string, error) {
	pathSeparator := string(os.PathSeparator)
	var volLen int
	// 当path是绝对路径时
	if volLen < len(path) && os.IsPathSeparator(path[volLen]) {
		volLen++
	}
	vol := path[:volLen]
	dest := path[:volLen]
	linksWalked := 0

	// 无论path是否为abs都可以从0开始
	for start, end := volLen, volLen; start < len(path); start = end {
		// 寻找start作为next subPath的首索引. start会跳过分隔符
		for start < len(path) && os.IsPathSeparator(path[start]) {
			start++
		}
		// end-1为next subPath的尾索引
		end = start
		for end < len(path) && !os.IsPathSeparator(path[end]) {
			end++
		}

		// 子path：path[start:end]. if中的都为特殊处理
		if start == end {
			// 没有剩余的path了，跳出
			break
		} else if path[start:end] == "." {
			// 忽略当前路径'.'
			continue
		} else if path[start:end] == ".." {
			// 回退一个子路径
			var r int
			for r = len(dest) - 1; r >= volLen; r-- {
				if os.IsPathSeparator(dest[r]) {
					break
				}
			}

			// TODO: 不明白这个逻辑？
			if r < volLen || dest[r+1:] == ".." {
				// Either path has no slashes
				// (it's empty or just "C:")
				// or it ends in a ".." we had to keep.
				// Either way, keep this "..".
				if len(dest) > volLen {
					dest += pathSeparator
				}
				dest += ".."
			} else {
				// 后退
				dest = dest[:r]
			}
			continue
		}

		// 正常子path处理逻辑
		// path[start:end] 不包含路径分隔符
		// 由于start会跳过分隔符，在拼接路径时需要补充上。如果dest末尾不是分隔符，则添加
		if len(dest) > 0 && !os.IsPathSeparator(dest[len(dest)-1]) {
			dest += pathSeparator
		}
		dest += path[start:end]

		// 符号链接处理
		fi, err := os.Lstat(dest)
		if err != nil {
			return "", err
		}

		// 不是符号链接
		if fi.Mode()&os.ModeSymlink == 0 {
			// 异常情况处理
			if !fi.Mode().IsDir() && end < len(path) {
				return "", syscall.ENOTDIR
			}
			// fi是目录，或者fi是普通文件且path已经遍历完。正常
			continue
		}

		// 符号链接
		linksWalked++
		// 符号链接最大支持255次
		if linksWalked > 255 {
			return "", ErrTooManyLinks
		}
		link, err := os.Readlink(dest)
		if err != nil {
			return "", err
		}
		// 用符号链接更新整个path
		// path.end对应的是分隔符, path[end:]是剩余的路径
		path = link + path[end:]

		if len(link) > 0 && os.IsPathSeparator(link[0]) {
			// link是绝对路径
			// 从新开始
			dest = link[:1]
			end = 1
		} else {
			// 相对路径，dest回退一个路径
			var r int
			for r = len(dest) - 1; r >= volLen; r-- {
				if os.IsPathSeparator(dest[r]) {
					break
				}
			}
			// 回归初始状态
			if r < volLen {
				dest = vol
			} else {
				// 回退
				dest = dest[:r]
			}
			// 相对路径，从0开始
			end = 0
		}
	}
	return Clean(dest), nil
}
