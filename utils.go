package httpcache

func contains(slice []int, searchValue int) bool {
	for index := range slice {
		if searchValue == slice[index] {
			return true
		}
	}
	return false
}
