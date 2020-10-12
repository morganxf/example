package filepath

import (
	"errors"
	"os"
	"sort"
)

// 标识是否跳过本文件所在的目录，不继续扫描本文件所在的目录
var SkipDir = errors.New("skip this directory")

// 为了单元测试，可以采用monkey.Patch
var lstat = os.Lstat

// WalkFunc的作用是对遍历到的file or dir进行进一步的处理，比如：过滤
// 该函数由Walk使用方定义
type WalkFunc func(path string, info os.FileInfo, err error) error

// Walk 遍历root文件树，并调用walkFn对各个文件进行处理。
func Walk(root string, walkFn WalkFunc) error {
	info, err := lstat(root)
	if err != nil {
		err = walkFn(root, nil, err)
	} else {
		err = walk(root, info, walkFn)
	}
	// SkipDir不认为是一个错误
	if err == SkipDir {
		return nil
	}
	return err
}

// 递归遍历path
// 入参包含os.FileInfo的原因是为了减少os.Lstat调用次数
func walk(path string, info os.FileInfo, walkFn WalkFunc) error {
	if !info.IsDir() {
		return walkFn(path, info, nil)
	}

	names, err := readDirNames(path)
	if err != nil {
		return walkFn(path, info, err)
	}
	if err := walkFn(path, info, nil); err != nil {
		return err
	}

	for _, name := range names {
		filename := Join(path, name)
		fileInfo, err := lstat(filename)
		if err != nil {
			if err := walkFn(filename, fileInfo, err); err != nil && err != SkipDir {
				return err
			}
			// continue
		} else {
			if err := walk(filename, fileInfo, walkFn); err != nil {
				if err != SkipDir {
					return err
				}
				// if err==SkipDir && filename是文件而非目录的话，则skip本filename所在的dir
				if !fileInfo.IsDir() {
					// 此时return的是SkipDir error
					return err
				}
			}
			// continue
		}
	}
	return nil
}

func readDirNames(dirname string) ([]string, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	names, err := f.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	return names, nil
}
