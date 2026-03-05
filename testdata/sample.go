package sample

import "fmt"

func Add(a, b int) int {
	return a + b
}

func Compare(a, b int) bool {
	if a < b {
		return true
	} else {
		return false
	}
}

func Logic(a, b bool) bool {
	return a && b
}

func Counter(n int) int {
	total := 0
	for i := 0; i < n; i++ {
		total += i
	}
	return total
}

func Negative(x int) int {
	return -x
}

func SwitchCase(x int) string {
	switch x {
	case 1:
		return "one"
	case 2:
		return "two"
	default:
		return "other"
	}
}

func Loop(n int) {
	for i := 0; i < n; i++ {
		if i == 5 {
			break
		}
	}
}

func Validate(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("empty")
	}
	return len(s), nil
}

func IsValid(s string) bool {
	return len(s) > 0
}

func Lookup(key string) (string, bool) {
	if key == "found" {
		return "value", true
	}
	return "", false
}

func MustParse(s string) int {
	if s == "" {
		panic("empty input")
	}
	return len(s)
}

func PanicVoid() {
	panic("fatal")
}

func Concat(a, b string) string {
	return a + b
}

func FullName(first, last string) string {
	return Concat(first, last)
}

func SafeLen(s *string) int {
	if s != nil {
		return len(*s)
	}
	return 0
}

func Head(items []int) int {
	if len(items) == 0 {
		return -1
	}
	return items[0]
}

var Debug = true
