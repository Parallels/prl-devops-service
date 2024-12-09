package cache

var cachedModels = make(map[string]string)

const CacheObjectNotFound = "{ object }"

func Get(modelName string) string {
	if val, ok := cachedModels[modelName]; ok {
		return val
	}
	return CacheObjectNotFound
}

func Push(modelName, json string) {
	cachedModels[modelName] = json
}
