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
	// if the lowest installed version is 2 or greater, use version 1 of node
	// accoutns
	if version >= 2 {
		return 1
	}

	return 0 // default
}
