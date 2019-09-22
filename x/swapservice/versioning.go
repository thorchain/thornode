package swapservice

func getVersion(height int64, prefix dbPrefix) int {
	switch prefix {
	case prefixTrustAccount:
		return getTrustAccountVersion(height)
	default:
		return 0
	}
}

func getTrustAccountVersion(height int64) int {
	return 0
}
