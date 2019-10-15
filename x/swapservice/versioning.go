package swapservice

func getVersion(height int64, prefix dbPrefix) int {
	switch prefix {
	case prefixNodeAccount:
		return getNodeAccountVersion(height)
	default:
		return 0
	}
}

func getNodeAccountVersion(height int64) int {
	return 0
}
