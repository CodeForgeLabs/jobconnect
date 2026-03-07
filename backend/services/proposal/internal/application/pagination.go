package application

import (
	"fmt"
	"strconv"
	"strings"
)

func normalizePageSize(v int32, fallback int) int {
	if v <= 0 {
		return fallback
	}
	if v > 100 {
		return 100
	}
	return int(v)
}

func parsePageToken(token string) (int, error) {
	if strings.TrimSpace(token) == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(token)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("invalid page_token")
	}
	return n, nil
}

func nextPageToken(currentOffset, pageSize, itemCount int) string {
	if itemCount < pageSize {
		return ""
	}
	return strconv.Itoa(currentOffset + itemCount)
}
