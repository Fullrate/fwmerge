package main

import (
	"encoding/json"
	"sort"
)

type Ruleset map[string]Table

type Table map[string]*Chain

type Chain struct {
	Unmanaged bool
	Policy    string
	Rules     RuleList
}

type RuleList []Rule

type Rule struct {
	Priority int
	Comment  string
	Content  string
}

func (r Ruleset) Merge(other Ruleset) {
	for t, cs := range other {
		if _, ok := r[t]; !ok {
			r[t] = make(Table)
		}
		tc := r[t]
		for c, cc := range cs {
			if _, ok := tc[c]; !ok {
				tc[c] = cc
			} else {
				tc[c].Rules = append(tc[c].Rules, cc.Rules...)
				if cc.Policy != "" {
					tc[c].Policy = cc.Policy
				}
			}
		}
	}
}

func (r Ruleset) Dump() string {
	b, _ := json.MarshalIndent(&r, "", "  ")
	return string(b)
}

func (r Ruleset) Sort() {
	for _, ts := range r {
		for _, cs := range ts {
			cs.Rules.Sort()
		}
	}
}

func (rl RuleList) Sort() {
	sort.Stable(rl)
}

func (rl RuleList) Len() int {
	return len(rl)
}

func (rl RuleList) Less(i, j int) bool {
	return rl[i].Priority < rl[j].Priority
}

func (rl RuleList) Swap(i, j int) {
	rl[i], rl[j] = rl[j], rl[i]
}
