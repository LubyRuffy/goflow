package gocodefuncs

import "sync"

// 全局object的管理对象，进行统一管理，后续要在调用者哪里进行配置通知
var globalObjects sync.Map

func registerObject(name string, description string) {
	globalObjects.Store(name, description)
}

// EachObjects 遍历注册的object（需要配置）
func EachObjects(f func(key, value string) bool) {
	globalObjects.Range(func(key, value interface{}) bool {
		return f(key.(string), value.(string))
	})
}
