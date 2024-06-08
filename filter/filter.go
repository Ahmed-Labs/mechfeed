package filter

import (
	"regexp"
	"strings"
)


func FilterKeywords(content string, keywords []string) bool {
	for _, keyword := range keywords {
		split_keywords := strings.Split(strings.ReplaceAll(keyword, " ", ""), ",")
		keyword_matched := true

		for _, k := range split_keywords {
			if len(k) == 0 {
				keyword_matched = false
				break
			}
			var re string = "(?i)"
			var filter_condition bool

			if k[0] == '-' && len(k) > 1 {
				re += regexp.QuoteMeta(k[1:])
				filter_condition = true
			} else {
				re += regexp.QuoteMeta(k)
				filter_condition = false
			}
			c := regexp.MustCompile(re)

			if c.MatchString(content) == filter_condition {
				keyword_matched = false
				break
			}
		}
		if keyword_matched {
			return true
		}
	}
	return false
}