package sonar

import "github.com/RandomCodeSpace/qctx/internal/model"

type ruleResp struct {
	Rule struct {
		Key      string `json:"key"`
		HTMLDesc string `json:"htmlDesc"`
	} `json:"rule"`
}

// GetRuleHTML returns rule.htmlDesc for the key; cached after first fetch.
func (c *Client) GetRuleHTML(key string) (string, error) {
	c.rulesMu.Lock()
	if c.rulesCache == nil {
		c.rulesCache = make(map[string]string)
	}
	if v, ok := c.rulesCache[key]; ok {
		c.rulesMu.Unlock()
		return v, nil
	}
	c.rulesMu.Unlock()

	var r ruleResp
	if err := c.GetJSON("/api/rules/show", map[string]string{"key": key}, &r); err != nil {
		return "", err
	}
	c.rulesMu.Lock()
	c.rulesCache[key] = r.Rule.HTMLDesc
	c.rulesMu.Unlock()
	return r.Rule.HTMLDesc, nil
}

// EnrichIssueDescriptions populates RuleDescHTML for each unique rule across the slice.
func (c *Client) EnrichIssueDescriptions(issues []model.Issue) {
	seen := map[string]string{}
	for i := range issues {
		key := issues[i].Rule
		if key == "" {
			continue
		}
		desc, ok := seen[key]
		if !ok {
			d, err := c.GetRuleHTML(key)
			if err != nil {
				continue
			}
			desc = d
			seen[key] = d
		}
		issues[i].RuleDescHTML = desc
	}
}

// EnrichHotspotDescriptions populates RuleDescHTML for each unique rule across the slice.
func (c *Client) EnrichHotspotDescriptions(hs []model.Hotspot) {
	seen := map[string]string{}
	for i := range hs {
		key := hs[i].Rule
		if key == "" {
			continue
		}
		desc, ok := seen[key]
		if !ok {
			d, err := c.GetRuleHTML(key)
			if err != nil {
				continue
			}
			desc = d
			seen[key] = d
		}
		hs[i].RuleDescHTML = desc
	}
}
