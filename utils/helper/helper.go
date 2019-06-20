package helper

import (
	"os/exec"
	"reflect"
)

// InArray将搜索数组中任何类型的元素。
// 将返回匹配元素的布尔值和索引。
// 如果元素存在，则 返回True,以及大于0的索引,
func InArray(needle interface{}, haystack interface{}) (exists bool, index int) {
	exists = false
	index = -1

	switch reflect.TypeOf(haystack).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(haystack)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(needle, s.Index(i).Interface()) == true {
				index = i
				exists = true
				return
			}
		}
	}

	return
}

// 处理 合并
func HandlePull(path, remoteName, ref string) error {
	var _cmd *exec.Cmd
	var _err error

	// 强制 fetch
	_cmd = exec.Command("git", "fetch", "--all")
	_cmd.Dir = path
	_err = _cmd.Run()

	if _err != nil {
		return _err
	}
	// 强制 reset
	_cmd = exec.Command("git", "reset", "--hard", remoteName+"/"+ref)
	_cmd.Dir = path
	_err = _cmd.Run()

	if _err != nil {
		return _err
	}
	// 强制 fetch
	_cmd = exec.Command("git", "pull")
	_cmd.Dir = path
	_err = _cmd.Run()

	if _err != nil {
		return _err
	}
	return nil
}
