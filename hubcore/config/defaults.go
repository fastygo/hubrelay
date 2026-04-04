package config

import (
	"strconv"
	"strings"
)

func DefaultString(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func DefaultStringAny(keys map[string]string, fallback string) string {
	for _, value := range keys {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return fallback
}

func DefaultStringFromEnv(values []string, fallback string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return fallback
}

func DefaultBoolFromEnv(values []string, fallback bool) bool {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return fallback
		}
		return parsed
	}
	return fallback
}
