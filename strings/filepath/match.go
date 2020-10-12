package filepath

import (
	"errors"
	"os"
	"strings"
	"unicode/utf8"
)

var ErrBadPattern = errors.New("syntax error in pattern")

// Match 匹配
//
// '/'分割符需要特殊处理，不能用'*'匹配
// '*'	匹配任意非分隔符的字符
// '?'  匹配任意一个非分割符的字符
// []	范围匹配，内不支持'*', '?'的上述作用
// \\	转义字符
//
func Match(pattern, name string) (matched bool, err error) {
Pattern:
	for len(pattern) > 0 {
		var startWithStar bool
		var chunk string
		startWithStar, chunk, pattern = scanChunk(pattern)
		// 说明pattern只剩下'*'来匹配剩余的name
		if startWithStar && chunk == "" {
			return !strings.Contains(name, string(os.PathSeparator)), nil
		}
		// 非'*'开头，只能从name起始处匹配
		if !startWithStar {
			restName, matched, err := matchChunk(chunk, name)
			if err != nil {
				return false, err
			}
			if !matched {
				return false, nil
			}
			name = restName
		} else {
			// 以'*'起始，非贪婪匹配
			for i := 0; i < len(name); i++ {
				// 由于'/'需要特殊处理，不能用'*'匹配。所以当i-1为'/'时，无法继续for循环
				if i-1 >= 0 && os.IsPathSeparator(name[i-1]) {
					break
				}
				restName, ok, err := matchChunk(chunk, name[i:])
				if ok {
					// chunk是最后一个subPattern, 且name仍有剩余. 则需要继续match检测
					if len(pattern) == 0 && len(restName) > 0 {
						continue
					}
					// 匹配成功，且不满足上述条件，继续剩余的pattern匹配
					name = restName
					continue Pattern
				}
				if err != nil {
					return false, nil
				}
				// 如果不匹配且没有错误，则i+=1再次match检测
			}
			// 无法匹配
			return false, nil
		}
	}
	// name匹配没有残留
	return len(name) == 0, nil
}

// scanChunk 迭代解析pattern，以非首'*'字符为分隔符，分割pattern为第一部分chunk和剩余的部分rest。
// chunk作为match的一个最小单位, 以'*'为分割符从pattern分割
func scanChunk(pattern string) (startWithStar bool, chunk string, rest string) {
	// 去除多余的连续重复的'*'
	for len(pattern) > 0 && pattern[0] == '*' {
		pattern = pattern[1:]
		startWithStar = true
	}
	inrange := false
	var i int
Scan:
	for i = 0; i < len(pattern); i++ {
		switch pattern[i] {
		// 转义符号后面的字符需直接跳过，i++
		case '\\':
			// i作为chunk和rest的分割索引，最大的合法值是len(pattern)
			// 当i==len(pattern)-1时，不做跳过
			if i+1 < len(pattern) {
				i++
			}
		case '[':
			inrange = true
		case ']':
			inrange = false
		case '*':
			// "[]"里面的'*'不能作为chunk和rest的分割，因为'['和']'必须成对
			if !inrange {
				break Scan
			}
		}
	}
	return startWithStar, pattern[:i], pattern[i:]
}

// matchChunk 检测chunk是否匹配s的起始部分
func matchChunk(chunk, s string) (rest string, matched bool, err error) {
	for len(chunk) > 0 {
		if len(s) == 0 {
			return
		}
		switch chunk[0] {
		// 范围匹配是指s中的一个字符在[]表示的范围之内
		// 1. 范围 x-y
		// 2. 枚举 abc
		case '[':
			// 获取s中的字符r
			r, n := utf8.DecodeRuneInString(s)
			s = s[n:]
			chunk = chunk[1:]
			if len(chunk) == 0 {
				err = ErrBadPattern
				return
			}
			negated := chunk[0] == '^'
			if negated {
				chunk = chunk[1:]
			}
			match := false
			nrange := 0
			for {
				// range匹配结束
				// 当nrange==0, 即chunk=="[]"时，此时ErrBadPattern。所以需要加上"nrange > 0"条件
				// 范围匹配模式匹配']'必须加转义，但是匹配'[', '*', '?'是可以不加转义的。
				// 当前match在范围模式中是不支持'*', '?'匹配语义的
				if len(chunk) > 0 && chunk[0] == ']' && nrange > 0 {
					chunk = chunk[1:]
					break
				}
				var lo, hi rune
				if lo, chunk, err = getEcs(chunk); err != nil {
					return
				}
				hi = lo
				if chunk[0] == '-' {
					if hi, chunk, err = getEcs(chunk[1:]); err != nil {
						return
					}
				}
				if lo <= r && r <= hi {
					match = true
				}
				nrange++
			}
			if match == negated {
				return
			}

		case '?':
			if s[0] == os.PathSeparator {
				return
			}
			// ? 匹配一个unicode而不是一个byte
			_, n := utf8.DecodeRuneInString(s)
			s = s[n:]
			chunk = chunk[1:]
		//	转义匹配，转义语法检验
		case '\\':
			chunk = chunk[1:]
			if len(chunk) == 0 {
				err = ErrBadPattern
				return
			}
			// 转义符后字符检测
			fallthrough
		default:
			if chunk[0] != s[0] {
				return
			}
			// 继续向后迭代解析
			s = s[1:]
			chunk = chunk[1:]
		}
	}
	return s, true, nil
}

// getEcs 从取值范围中后去第一个合法字符
func getEcs(chunk string) (r rune, nchunk string, err error) {
	// 合法性检测
	if len(chunk) == 0 || chunk[0] == '-' || chunk[0] == ']' {
		err = ErrBadPattern
		return
	}
	// 转义处理
	if chunk[0] == '\\' {
		chunk = chunk[1:]
		if len(chunk) == 0 {
			err = ErrBadPattern
			return
		}
	}
	r, n := utf8.DecodeRuneInString(chunk)
	if r == utf8.RuneError && n == 1 {
		err = ErrBadPattern
	}
	nchunk = chunk[n:]
	if len(nchunk) == 0 {
		err = ErrBadPattern
	}
	return
}
