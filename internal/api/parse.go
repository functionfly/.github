package api

import (
	"strconv"
)

func ParseIntParam(s string, defaultVal int) int {
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return defaultVal
}

func ParseIntParamWithMin(s string, defaultVal, min int) int {
	if n, err := strconv.Atoi(s); err == nil && n >= min {
		return n
	}
	return defaultVal
}

func ParseIntParamWithRange(s string, defaultVal, min, max int) int {
	if n, err := strconv.Atoi(s); err == nil && n >= min && n <= max {
		return n
	}
	return defaultVal
}