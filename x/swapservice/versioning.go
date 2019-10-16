package swapservice

func getVersion(sversion int, prefix dbPrefix) int {
	switch prefix {
	case prefixNodeAccount:
		return getNodeAccountVersion(sversion)
	default:
		return 0 // default
	}
}

func getNodeAccountVersion(sversion int) int {
	return 0 // default
}
