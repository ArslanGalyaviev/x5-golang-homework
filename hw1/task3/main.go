package main

func CheckCategory(sum int) string {
	switch {
	case sum < 0:
		return "invalid"
	case sum < 10000:
		return "small"
	case sum < 50000:
		return "regular"
	default:
		return "large"
	}
}
