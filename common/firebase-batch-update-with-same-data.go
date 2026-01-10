package common

func FirebaseBatchUpdateWithSameData(docPaths []string, updateData map[string]interface{}) error {
	if len(docPaths) == 0 {
		return nil
	}

	docPathMap := make(map[string]map[string]interface{}, len(docPaths))
	for _, path := range docPaths {
		docPathMap[path] = updateData
	}

	return FirebaseBatchUpdate(docPathMap)
}
