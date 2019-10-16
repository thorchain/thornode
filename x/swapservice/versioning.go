package swapservice

func getVersion(version int, prefix dbPrefix) int {
	switch prefix {
	case prefixNodeAccount:
		return getNodeAccountVersion(version)
	default:
		return 0 // default
	}
}

func getNodeAccountVersion(version int) int {
	return 0 // default
}
