package i18n

import "fmt"

// fmtSprintf is exposed via package-private alias so tests in i18n_test can
// stub it without touching fmt directly.
func fmtSprintf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}
