package utils

import (
    "strconv"
    "strings"
)

func ParseMemory(s string) (int64, error) {
    s = strings.TrimSpace(s)
    var multiplier int64 = 1
    if strings.HasSuffix(s, "g") || strings.HasSuffix(s, "G") {
        multiplier = 1024 * 1024 * 1024
        s = s[:len(s)-1]
    } else if strings.HasSuffix(s, "m") || strings.HasSuffix(s, "M") {
        multiplier = 1024 * 1024
        s = s[:len(s)-1]
    } else if strings.HasSuffix(s, "k") || strings.HasSuffix(s, "K") {
        multiplier = 1024
        s = s[:len(s)-1]
    } else {
        multiplier = 1
    }
    value, err := strconv.ParseFloat(s, 64)
    if err != nil {
        return 0, err
    }
    return int64(value * float64(multiplier)), nil
}
