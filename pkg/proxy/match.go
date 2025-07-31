package proxy

import (
	"net/http"

	"github.com/webbgeorge/local-alb/pkg/config"
)

func match(conf config.Config, r *http.Request) ([]config.Action, error) {
	for _, rule := range conf.Rules {
		matched, err := matchRule(rule, r)
		if err != nil {
			return nil, err
		}
		if matched {
			return rule.Actions, err
		}
	}
	return nil, nil
}

func matchRule(rule config.Rule, r *http.Request) (bool, error) {
	matched := true
	for _, condition := range rule.Conditions {
		m, err := matchCondition(condition, r)
		if err != nil {
			return false, err
		}
		if !m {
			matched = false
			break
		}
	}
	return matched, nil
}

func matchCondition(condition config.Condition, r *http.Request) (bool, error) {
	// TODO determine type
	// TODO implement real matchers
	if r.URL.Path == condition.PathPattern {
		return true, nil
	}
	return false, nil
}
