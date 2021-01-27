package util

func IsInStringList(list []string, s string) bool {
	for i := range list {
		if list[i] == s {
			return true
		}
	}
	return false
}
