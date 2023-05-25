package common

type Dictionary struct {
	Data map[string]interface{} `json:"data"`
}

func NewDictionary() *Dictionary {
	return &Dictionary{
		Data: make(map[string]interface{}),
	}
}

func (d *Dictionary) Add(keys []string, value interface{}) {
	lastKeyIndex := len(keys) - 1
	currentDict := d.Data

	for i, key := range keys {
		if i == lastKeyIndex {
			currentDict[key] = value
		} else {
			if nestedDict, ok := currentDict[key].(map[string]interface{}); !ok {
				nestedDict = make(map[string]interface{})
				currentDict[key] = nestedDict
				currentDict = nestedDict
			} else {
				currentDict = nestedDict
			}
		}
	}
}

func (d *Dictionary) Get(keys []string) (interface{}, bool) {
	lastKeyIndex := len(keys) - 1
	currentDict := d.Data

	for i, key := range keys {
		if i == lastKeyIndex {
			value, exists := currentDict[key]
			return value, exists
		} else {
			if nestedDict, ok := currentDict[key].(map[string]interface{}); ok {
				currentDict = nestedDict
			} else {
				return nil, false
			}
		}
	}

	return nil, false
}

func (d *Dictionary) Remove(keys []string) {
	lastKeyIndex := len(keys) - 1
	currentDict := d.Data

	for i, key := range keys {
		if i == lastKeyIndex {
			delete(currentDict, key)
		} else {
			if nestedDict, ok := currentDict[key].(map[string]interface{}); ok {
				currentDict = nestedDict
			} else {
				return
			}
		}
	}
}
