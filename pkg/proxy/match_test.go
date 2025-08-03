package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/webbgeorge/local-alb/pkg/config"
)

func TestMatch(t *testing.T) {
	testCases := map[string]struct {
		conf            config.Config
		request         *http.Request
		expectedActions []config.Action
	}{
		"noRulesReturnsNil": {
			conf:            config.Config{},
			request:         httptest.NewRequest("GET", "/abc", nil),
			expectedActions: nil,
		},
		"singleRuleWithMatchReturnsActions": {
			conf: config.Config{Rules: []config.Rule{
				{
					Conditions: []config.Condition{
						{HTTPRequestMethod: &config.HTTPRequestMethodCondition{Values: []string{"GET"}}},
					},
					Actions: []config.Action{{Type: "TEST"}},
				},
			}},
			request:         httptest.NewRequest("GET", "/abc", nil),
			expectedActions: []config.Action{{Type: "TEST"}},
		},
		"singleRuleWithNoMatchReturnsNil": {
			conf: config.Config{Rules: []config.Rule{
				{
					Conditions: []config.Condition{
						{HTTPRequestMethod: &config.HTTPRequestMethodCondition{Values: []string{"POST"}}},
					},
					Actions: []config.Action{{Type: "TEST"}},
				},
			}},
			request:         httptest.NewRequest("GET", "/abc", nil),
			expectedActions: nil,
		},
		"multipleRulesWithNoMatchReturnsNil": {
			conf: config.Config{Rules: []config.Rule{
				{
					Conditions: []config.Condition{
						{HTTPRequestMethod: &config.HTTPRequestMethodCondition{Values: []string{"POST"}}},
					},
					Actions: []config.Action{{Type: "TEST"}},
				},
				{
					Conditions: []config.Condition{
						{PathPattern: &config.PathPatternCondition{Values: []string{"/xyz"}}},
					},
					Actions: []config.Action{{Type: "TEST"}},
				},
			}},
			request:         httptest.NewRequest("GET", "/abc", nil),
			expectedActions: nil,
		},
		"multipleRulesWithOneMatchReturnsActions": {
			conf: config.Config{Rules: []config.Rule{
				{
					Conditions: []config.Condition{
						{HTTPRequestMethod: &config.HTTPRequestMethodCondition{Values: []string{"POST"}}},
					},
					Actions: []config.Action{{Type: "TEST1"}},
				},
				{
					Conditions: []config.Condition{
						{PathPattern: &config.PathPatternCondition{Values: []string{"/abc"}}},
					},
					Actions: []config.Action{{Type: "TEST2"}},
				},
			}},
			request:         httptest.NewRequest("GET", "/abc", nil),
			expectedActions: []config.Action{{Type: "TEST2"}},
		},
		"multipleRulesWithManyMatchesReturnsFirstOne": {
			conf: config.Config{Rules: []config.Rule{
				{
					Conditions: []config.Condition{
						{HTTPRequestMethod: &config.HTTPRequestMethodCondition{Values: []string{"GET"}}},
					},
					Actions: []config.Action{{Type: "TEST1"}},
				},
				{
					Conditions: []config.Condition{
						{PathPattern: &config.PathPatternCondition{Values: []string{"/abc"}}},
					},
					Actions: []config.Action{{Type: "TEST2"}},
				},
			}},
			request:         httptest.NewRequest("GET", "/abc", nil),
			expectedActions: []config.Action{{Type: "TEST1"}},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			actions := match(tc.conf, tc.request)
			assert.Equal(t, tc.expectedActions, actions)
		})
	}
}

func TestMatchRule(t *testing.T) {
	testCases := map[string]struct {
		rule           config.Rule
		request        *http.Request
		expectedResult bool
	}{
		"manyConditionsWithNoMatchesReturnsFalse": {
			rule: config.Rule{Conditions: []config.Condition{
				{HTTPRequestMethod: &config.HTTPRequestMethodCondition{Values: []string{"POST"}}},
				{PathPattern: &config.PathPatternCondition{Values: []string{"/xyz"}}},
			}},
			request:        httptest.NewRequest("GET", "/abc", nil),
			expectedResult: false,
		},
		"manyConditionsWithOneMatchReturnsFalse": {
			rule: config.Rule{Conditions: []config.Condition{
				{HTTPRequestMethod: &config.HTTPRequestMethodCondition{Values: []string{"GET"}}},
				{PathPattern: &config.PathPatternCondition{Values: []string{"/xyz"}}},
			}},
			request:        httptest.NewRequest("GET", "/abc", nil),
			expectedResult: false,
		},
		"manyConditionsWithAllMatchesReturnsTrue": {
			rule: config.Rule{Conditions: []config.Condition{
				{HTTPRequestMethod: &config.HTTPRequestMethodCondition{Values: []string{"GET"}}},
				{PathPattern: &config.PathPatternCondition{Values: []string{"/abc"}}},
			}},
			request:        httptest.NewRequest("GET", "/abc", nil),
			expectedResult: true,
		},
		"singleConditionWithMatchReturnsTrue": {
			rule: config.Rule{Conditions: []config.Condition{
				{HTTPRequestMethod: &config.HTTPRequestMethodCondition{Values: []string{"GET"}}},
			}},
			request:        httptest.NewRequest("GET", "/abc", nil),
			expectedResult: true,
		},
		"singleConditionWithNoMatchReturnsFalse": {
			rule: config.Rule{Conditions: []config.Condition{
				{HTTPRequestMethod: &config.HTTPRequestMethodCondition{Values: []string{"POST"}}},
			}},
			request:        httptest.NewRequest("GET", "/abc", nil),
			expectedResult: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			expectedResult := matchRule(tc.rule, tc.request)
			assert.Equal(t, tc.expectedResult, expectedResult)
		})
	}
}

// TODO test individual matcher functions
