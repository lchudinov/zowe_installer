package launcher

import "regexp"

func stripEscapeSeqs(s string) string {
	re := regexp.MustCompile(`\x1b\[(0;)*\d{2}m`)
	return re.ReplaceAllString(s, "")
}
