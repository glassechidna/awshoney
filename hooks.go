package awshoney

import "sync"

var cachedMap map[string]string
var loadCache sync.Once

func PresendHook(m map[string]interface{}) {
	loadCache.Do(func() {
		cachedMap = Map()
	})

	for k, v := range cachedMap {
		m[k] = v
	}
}

func ComposePresendHooks(funcs ...func(map[string]interface{})) func(map[string]interface{}) {
	return func(m map[string]interface{}) {
		for _, f := range funcs {
			f(m)
		}
	}
}
