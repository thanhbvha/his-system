package logger

import (
	"regexp"
)

// PII mask regex patterns
var (
	phoneRe    = regexp.MustCompile(`(0[3|5|7|8|9])+([0-9]{8})\b`)
	emailRe    = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	cccdRe     = regexp.MustCompile(`\b([0-9]{12})\b`)
	passwordRe = regexp.MustCompile(`(?i)"password"\s*:\s*"[^"]*"`)
)

func MaskPII(logData string) string {
	res := phoneRe.ReplaceAllString(logData, "***MASKED_PHONE***")
	res = emailRe.ReplaceAllString(res, "***MASKED_EMAIL***")
	res = cccdRe.ReplaceAllString(res, "***MASKED_CCCD***")
	res = passwordRe.ReplaceAllString(res, `"password":"***MASKED***"`)
	return res
}
