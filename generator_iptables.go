package main

import (
	"fmt"
	"strings"
)

var POLICY_TABLES = map[string]map[string]struct{}{
	"filter": map[string]struct{}{
		"INPUT":   struct{}{},
		"FORWARD": struct{}{},
		"OUTPUT":  struct{}{},
	},
	"nat": map[string]struct{}{
		"PREROUTING":  struct{}{},
		"INPUT":       struct{}{},
		"OUTPUT":      struct{}{},
		"POSTROUTING": struct{}{},
	},
	"mangle": map[string]struct{}{
		"PREROUTING":  struct{}{},
		"INPUT":       struct{}{},
		"FORWARD":     struct{}{},
		"OUTPUT":      struct{}{},
		"POSTROUTING": struct{}{},
	},
	"raw": map[string]struct{}{
		"PREROUTING": struct{}{},
		"OUTPUT":     struct{}{},
	},
}

func generateTable(name string, t Table) (string, error) {
	var tmp string
	tmp += fmt.Sprintf("*%s\n", name)

	// Define chains
	for cname, chain := range t {
		policy := "-"
		if POLICY_TABLES[name] != nil {
			if _, ok := POLICY_TABLES[name][cname]; ok {
				policy = strings.ToUpper(chain.Policy)
				if policy == "" {
					policy = "ACCEPT"
				}
			}
		}
		tmp += fmt.Sprintf(":%s %s [0:0]\n", cname, policy)
	}

	// Clear chains that are not unmanaged
	for cname, chain := range t {
		if chain.Unmanaged {
			continue
		}
		tmp += fmt.Sprintf("-F %s\n", cname)
	}

	// Insert rules into chains
	for cname, chain := range t {
		if chain.Unmanaged {
			continue
		}
		for _, r := range chain.Rules {
			tmp += fmt.Sprintf("-A %s %s\n", cname, r.Content)
		}
	}

	tmp += "COMMIT\n"
	return tmp, nil
}

func GenerateIptables(rs Ruleset) (string, error) {
	var tmp string
	for tname, tbl := range rs {
		td, err := generateTable(tname, tbl)
		if err != nil {
			return "", err
		}
		tmp += td
	}
	return tmp, nil
}
