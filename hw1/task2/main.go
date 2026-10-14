package main

import "fmt"

func CalculateStats(arr []int) (int, int, int, float64) {
	if len(arr) == 0 {
		return 0, 0, 0, 0
	}
	minVal := arr[0]
	maxVal := arr[0]
	sum := 0
	for _, v := range arr {
		minVal = min(minVal, v)
		maxVal = max(maxVal, v)
		sum += v
	}
	avg := float64(sum) / float64(len(arr))
	return minVal, maxVal, sum, avg
}

func main() {
	arr := []int{12000, 7500, 31000, 4500}
	min, max, sum, avg := CalculateStats(arr)
	fmt.Printf("Min: %d, Max: %d, Sum: %d, Avg: %.2f\n", min, max, sum, avg)
}
