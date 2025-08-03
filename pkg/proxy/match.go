package proxy

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gobwas/glob"
	"github.com/webbgeorge/local-alb/pkg/config"
)

// return actions for first matching rule
func match(conf config.Config, r *http.Request) []config.Action {
	for _, rule := range conf.Rules {
		if matchRule(rule, r) {
			return rule.Actions
		}
	}
	return nil
}

// returns true if all conditions of rule are met
func matchRule(rule config.Rule, r *http.Request) bool {
	for _, condition := range rule.Conditions {
		if !matchCondition(condition, r) {
			return false
		}
	}
	return true
}

func matchCondition(condition config.Condition, r *http.Request) bool {
	if condition.HostHeader != nil {
		return matchHostHeader(condition, r)
	} else if condition.PathPattern != nil {
		return matchPathPattern(condition, r)
	} else if condition.QueryString != nil {
		return matchQueryString(condition, r)
	} else if condition.HTTPHeader != nil {
		return matchHTTPHeader(condition, r)
	} else if condition.HTTPRequestMethod != nil {
		return matchHTTPRequestMethod(condition, r)
	}
	return false
}

// TODO could optimise match functions by compiling glob patterns at startup

func matchHostHeader(condition config.Condition, r *http.Request) bool {
	for _, val := range condition.HostHeader.Values {
		g := glob.MustCompile(strings.ToLower(val))
		if g.Match(strings.ToLower(r.Host)) {
			return true
		}
	}
	return false
}

func matchPathPattern(condition config.Condition, r *http.Request) bool {
	for _, val := range condition.PathPattern.Values {
		g := glob.MustCompile(strings.ToLower(val))
		if g.Match(strings.ToLower(r.URL.Path)) {
			return true
		}
	}
	return false
}

func matchQueryString(condition config.Condition, r *http.Request) bool {
	for _, val := range condition.QueryString.Values {
		g := glob.MustCompile(val.Value)
		if g.Match(r.URL.Query().Get(val.Key)) {
			return true
		}
	}
	return false
}

func matchHTTPHeader(condition config.Condition, r *http.Request) bool {
	for _, val := range condition.HTTPHeader.Values {
		g := glob.MustCompile(val)
		if g.Match(r.Header.Get(condition.HTTPHeader.HTTPHeaderName)) {
			return true
		}
	}
	return false
}

func matchHTTPRequestMethod(condition config.Condition, r *http.Request) bool {
	return slices.Contains(condition.HTTPRequestMethod.Values, r.Method)
}
