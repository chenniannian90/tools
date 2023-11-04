package reflect

import (
	"strconv"
	"strings"
)

func ParseStructTags(tag string) map[string]StructTag {
	tagFlags := map[string]StructTag{}

	for tag != "" {
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		name := string(tag[:i])
		tag = tag[i+1:]

		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		qvalue := tag[:i+1]
		tag = tag[i+1:]

		value, err := strconv.Unquote(qvalue)
		if err != nil {
			break
		}
		tagFlags[name] = StructTag(value)
	}

	return tagFlags
}

type StructTag string

func (t StructTag) Name() string {
	s := string(t)

	if i := strings.Index(s, ","); i >= 0 {
		if i == 0 {
			return ""
		}
		return s[0:i]
	}

	return s
}

func (t StructTag) HasFlag(flag string) bool {
	if i := strings.Index(string(t), flag); i > 0 {
		return true
	}
	return false
}
