package filepath

import "os"

// 只考虑linux
const (
	Separator     = os.PathSeparator
	ListSeparator = os.PathListSeparator
)

// lazybuf 只有在append需要时才会分配新的空间存储path数据
type lazybuf struct {
	// path原始数据
	path string
	// path byte缓冲，只有在需要的时候才会分配空间
	buf []byte
	// buf中下一个写入的索引
	w int
}

func (b *lazybuf) index(i int) byte {
	if b.buf != nil {
		return b.buf[i]
	}
	return b.path[i]
}

func (b *lazybuf) append(c byte) {
	if b.buf == nil {
		if b.w < len(b.path) && b.path[b.w] == c {
			b.w++
			return
		}
		// 当要append的byte c与path对应位置的byte不相等时，需要分配新的空间来存储append数据
		// b.w最大不会超过len(b.path)
		b.buf = make([]byte, len(b.path))
		copy(b.buf, b.path[:b.w])
	}
	b.buf[b.w] = c
	b.w++
}

func (b *lazybuf) string() string {
	// 说明append结束时，append的数据正好是path的前w个
	if b.buf == nil {
		return b.path[:b.w]
	}
	return string(b.buf[:b.w])
}

// Clean清理path中多余的字符, 且如果不是根目录最后一个字符不会是分隔符
//
// 清理规则:
// 1. 去除path中多余分隔符Separator
// 2. 忽略path中的'./'目录。 因为'./'代表当前目录
// 3. 去除'../'及其之前的目录，如果'../'之前存在目录的话。因为'../'代表父目录
// 4. 如果'../'之前是根目录的话，直接去除。因为根目录之前没有目录
// 如果path为空的话，返回当前目录'.'
//
func Clean(path string) string {
	if path == "" {
		return "."
	}

	// 检测是否为绝对路径
	rooted := os.IsPathSeparator(path[0])

	pathLen := len(path)
	out := lazybuf{path: path}
	r, dotdot := 0, 0
	if rooted {
		// 首先append分隔符 并更新r index
		out.append(Separator)
		r, dotdot = 1, 1
	}

	// 遍历path，处理一个个subPath
	for r < pathLen {
		switch {
		// 当前byte为分隔符
		// 跳过
		case os.IsPathSeparator(path[r]):
			r++
		// 当前byte为'.'字符且为最后一个字符 || 当前byte为'.'字符且下一个字符为path分隔符
		// 忽略
		case path[r] == '.' && (r+1 == pathLen || os.IsPathSeparator(path[r+1])):
			r++
		//	既然能够走到这个case，就说明不满足上面的第二个条件。也就保证r+1<pathLen
		case path[r] == '.' && path[r+1] == '.' && (r+2 == pathLen || os.IsPathSeparator(path[r+2])):
			// 跳过两个字符
			r = r + 2
			// 回溯，如果上一个subPath存在 则删除
			switch {
			// 能够回溯
			case out.w > dotdot:
				// 指向当前out中最后一个元素的位置
				out.w--
				// append数据回退，直到分隔符。此时out.w指向最后一个分割符，如果存在
				for out.w > dotdot && !os.IsPathSeparator(out.index(out.w)) {
					out.w--
				}
			// out.w == dotdot
			// 不能回溯且path为绝对路径，则已经达到根路径，不做任何处理。
			// 不能回溯且path为相对路径, 则说明已经
			case !rooted:
				// 相对路径模式，首字符不能为分隔符
				if out.w > 0 {
					out.append(Separator)
				}
				out.append('.')
				out.append('.')
				// 更新dotdot，因为相对路径模式不能删除前面的'..'子路径（绝对路径由于已经存在根，多余的'..'就被删除掉了）
				dotdot = out.w
			}
		// 当前subPath不为".", ".."
		default:
			// 新的subPath，前需要加分隔符
			// out.w != 1 || out.w != 0 分别代表 当不是第一个subPath时添加分割符。起始分隔符已经append到out
			if rooted && out.w != 1 || !rooted && out.w != 0 {
				out.append(Separator)
			}
			for ; r < pathLen && !os.IsPathSeparator(path[r]); r++ {
				out.append(path[r])
			}
		}
	}
	if out.w == 0 {
		out.append('.')
	}
	return out.string()
}
