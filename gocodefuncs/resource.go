package gocodefuncs

import "github.com/LubyRuffy/goflow/utils"

// GetRuntimeValue 在 defaultValue 为空时，获取Runner中的环境变量并返回
func GetRuntimeValue(p Runner, name, defaultValue string) string {
	if defaultValue == "" {
		if value, ok := p.GetObject(name); ok {
			return value.(string)
		}
	}
	return defaultValue
}

// UseGlobalValue 根据存储的key决定是否使用全局变量
func UseGlobalValue(p Runner, name string) bool {
	if value, ok := p.GetObject(name); ok {
		if use, ok := value.(bool); ok && use {
			return true
		}
	}
	return false
}

// AddResourceField 在object中添加资源字段
func AddResourceField(p Runner, field string) {
	AddObjectSlice(p, utils.ResourceFieldsObjectName, field)
}

// AddResource 在object中添加资源列表
func AddResource(p Runner, resource string) {
	AddObjectSlice(p, utils.ResourcesObjectName, resource)
}

// AddStaticResource 在object中添加static资源
func AddStaticResource(p Runner, resource string) {
	AddObjectSlice(p, utils.StaticResourceObjectName, resource)
}

// ReplaceResourcePath 替换 object 中的资源
func ReplaceResourcePath(p Runner, old, new string) {
	ReplaceObjectSlice(p, utils.ResourcesObjectName, old, new)
}

// ReplaceStaticResourcePath 替换 object 中的静态资源
func ReplaceStaticResourcePath(p Runner, old, new string) {
	ReplaceObjectSlice(p, utils.StaticResourceObjectName, old, new)
}

// AddObjectSlice 在object slice 中添加元素
func AddObjectSlice(p Runner, objectName, ele string) {
	var result []string
	if res, ok := p.GetObject(objectName); ok {
		if result, ok = res.([]string); !ok {
			result = []string{}
		}
	} else {
		result = []string{}
	}
	result = append(result, ele)
	p.SetObject(objectName, result)
}

// ReplaceObjectSlice 在 object slice 中替换元素
func ReplaceObjectSlice(p Runner, objectName, old, new string) {
	var result []string
	if res, ok := p.GetObject(objectName); ok {
		if result, ok = res.([]string); !ok {
			result = []string{}
		}
	} else {
		result = []string{}
	}

	for i, s := range result {
		if s == old {
			result[i] = new
		}
	}
	p.SetObject(objectName, result)
}
