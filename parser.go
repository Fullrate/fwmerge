package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strconv"
	"strings"
)

type parse_ruleset map[string]parse_table
type parse_table map[string]interface{}

func yamlToString(v interface{}) (string, bool) {
	switch v := v.(type) {
	case string:
		return v, true
	case int:
		return fmt.Sprintf("%d", v), true
	case int8:
		return fmt.Sprintf("%d", v), true
	case int16:
		return fmt.Sprintf("%d", v), true
	case int32:
		return fmt.Sprintf("%d", v), true
	case int64:
		return fmt.Sprintf("%d", v), true
	case float32:
		return fmt.Sprintf("%f", v), true
	case float64:
		return fmt.Sprintf("%f", v), true
	default:
		return "", false
	}
}

func stripLineBreaks(s string) string {
	s = strings.Replace(s, "\r\n", " ", -1)
	s = strings.Replace(s, "\r", " ", -1)
	s = strings.Replace(s, "\n", " ", -1)
	return s
}

func parsePriorityComment(key string) (prio int, comment string, err error) {
	pn := strings.SplitN(key, " ", 2)
	if len(pn) == 2 {
		comment = pn[1]
	}
	pstr := strings.TrimSpace(pn[0])
	if len(pstr) == 0 {
		err = fmt.Errorf("Empty key")
		return
	}
	prio64, err := strconv.ParseInt(pn[0], 10, 64)
	if err != nil {
		return
	}
	prio = int(prio64)
	return
}

func ParseString(content string) (Ruleset, error) {
	var prs parse_ruleset
	if err := yaml.Unmarshal([]byte(content), &prs); err != nil {
		return nil, err
	}

	rs := make(Ruleset)
	for t, cs := range prs {
		table := make(Table)
		for c, cc := range cs {
			var chain Chain

			switch cc := cc.(type) {
			case string:
				if strings.ToLower(cc) == "unmanaged" {
					chain.Unmanaged = true
				} else {
					return nil, fmt.Errorf("Unknown chain type '%s'", cc)
				}
			case []interface{}:
				for id, rule := range cc {
					rule, ok := rule.(map[interface{}]interface{})
					if !ok {
						return nil, fmt.Errorf("Invalid rule format in %s:%s, rule %d", t, c, id+1)
					}
					for k, v := range rule {
						k, ok := yamlToString(k)
						if !ok {
							return nil, fmt.Errorf("Invalid rule name in %s:%s, rule %d", t, c, id+1)
						}

						v, ok := yamlToString(v)
						if !ok {
							return nil, fmt.Errorf("Invalid rule content in %s:%s, rule %d(\"%s\")", t, c, id+1, k)
						}

						k = strings.TrimSpace(k)
						switch k {
						case "":
							return nil, fmt.Errorf("Empty key in %s:%s, rule %d", t, c, id+1)
						case "policy":
							chain.Policy = v
						default:
							prio, comment, err := parsePriorityComment(k)
							if err != nil {
								return nil, fmt.Errorf("Invalid priority in %s:%s, rule %d: %v", t, c, id+1, err)
							}
							chain.Rules = append(chain.Rules, Rule{prio, comment, stripLineBreaks(v)})
						}
						break // Skip all but the first rule
					}
				}
			default:
				return nil, fmt.Errorf("Could not parse chain %s:%s, wrong type", t, c)
			}

			table[c] = &chain
		}
		rs[t] = table
	}
	return rs, nil
}

func ParseFile(filename string) (Ruleset, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ParseString(string(buf))
}

func ParseFiles(filename []string) (Ruleset, error) {
	rs := make(Ruleset)
	for _, fn := range filename {
		frs, err := ParseFile(fn)
		if err != nil {
			return nil, fmt.Errorf("Error in file %s:\n%v", fn, err)
		}
		rs.Merge(frs)
	}
	return rs, nil
}
