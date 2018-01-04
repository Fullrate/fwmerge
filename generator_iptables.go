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

func chainHasPolicy(tname, cname string) bool {
	if POLICY_TABLES[tname] != nil {
		if _, ok := POLICY_TABLES[tname][cname]; ok {
			return true
		}
	}
	return false
}

func generateTable(name string, t Table, withChains bool) (string, error) {
	var tmp string
	tmp += fmt.Sprintf("*%s\n", name)

	if withChains {
		// Declare all chains. This will also flush them.
		for cname, chain := range t {
			policy := "-"
			if chainHasPolicy(name, cname) {
				policy = strings.ToUpper(chain.Policy)
				if policy == "" {
					policy = "ACCEPT"
				}
			}
			tmp += fmt.Sprintf(":%s %s [0:0]\n", cname, policy)
		}
	} else {
		// Clear chains that are not unmanaged
		for cname, chain := range t {
			if chain.Unmanaged {
				continue
			}
			tmp += fmt.Sprintf("-F %s\n", cname)
			if chainHasPolicy(name, cname) {
				policy := strings.ToUpper(chain.Policy)
				if policy == "" {
					policy = "ACCEPT"
				}
				tmp += fmt.Sprintf("-P %s %s\n", cname, policy)
			}
		}
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

func GenerateIptables(rs Ruleset, withChains bool) (string, error) {
	var tmp string
	for tname, tbl := range rs {
		td, err := generateTable(tname, tbl, withChains)
		if err != nil {
			return "", err
		}
		tmp += td
	}
	return tmp, nil
}

func GenerateIptablesChains(rs Ruleset) (string, error) {
	var tmp string
	for tname, tbl := range rs {
		for cname, _ := range tbl {
			// Default chains always exist, skip them
			if !chainHasPolicy(tname, cname) {
				tmp += fmt.Sprintf("%s %s\n", tname, cname)
			}
		}
	}
	return tmp, nil
}
