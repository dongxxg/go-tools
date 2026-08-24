package main

import "fmt"

func findMostFrequentIP(addresses []string) string {
	if len(addresses) == 1 {
		return addresses[0]
	}

	mid := len(addresses) / 2
	left := findMostFrequentIP(addresses[:mid])
	right := findMostFrequentIP(addresses[mid:])

	if countOccurrences(addresses, left) > mid {
		return left
	} else if countOccurrences(addresses, right) > mid {
		return right
	} else {
		return ""
	}
}

func countOccurrences(addresses []string, target string) int {
	count := 0
	for _, address := range addresses {
		if address == target {
			count++
		}
	}
	return count
}

func main() {
	// 假设有一个存储1亿条去重后的IPV4地址的切片
	addresses := []string{"192.168.0.1", "192.168.0.2", "192.168.0.3", "192.168.0.2", "192.168.0.2"}

	mostFrequentIP := findMostFrequentIP(addresses)
	fmt.Printf("The most frequent IP address is: %s\n", mostFrequentIP)
}
