package utils

// ContainsString checks if _s_ is in the array _slz_.
//
// If fails, it returns -1 and `false`.
func ContainsString(slz []string, s string) (int, bool) {

	for i := range slz {

		if slz[i] == s {
			return i, true
		}

	}

	return -1, false
}
