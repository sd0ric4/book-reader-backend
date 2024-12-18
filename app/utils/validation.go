package utils

import "regexp"

// ValidateEmail 用于验证邮箱格式
func ValidateEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$`)
	return re.MatchString(email)
}
