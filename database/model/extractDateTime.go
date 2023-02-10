package model

import (
	"fmt"
	"regexp"
)

var exp_data *regexp.Regexp

func init() {
	exp_data = regexp.MustCompile(`(?P<date>\d{4}-\d{2}-\d{2})(T|\s)(?P<time>\d{2}:\d{2}:\d{2}(\.\d+){0,1})`)
}

func ExtractDateTime(value string) (result map[string]string, err error) {
	match := exp_data.FindStringSubmatch(value)
	result = make(map[string]string)

	if len(match) == 0 {
		err = fmt.Errorf("string format is not valid: %s", value)
		return
	}

	for i, name := range exp_data.SubexpNames() {
		if i == 0 || name == "" {
			continue
		}
		result[name] = match[i]
	}
	return
}
