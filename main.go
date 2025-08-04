package main

import (
	"github.com/webbgeorge/local-alb/pkg/config"
	"github.com/webbgeorge/local-alb/pkg/proxy"
)

func main() {
	proxy.NewProxy(config.Config{
		Port: 8080,
		DefaultActions: []config.Action{{
			Type: "fixed-response",
			FixedResponse: config.FixedResponse{
				ContentType: "text/html",
				StatusCode:  200,
				MessageBody: "HELLO!",
			},
		}},
		Rules: []config.Rule{
			{
				Conditions: []config.Condition{{
					PathPattern: &config.PathPatternCondition{
						Values: []string{"/t?st/*/blue"},
					},
				}},
				Actions: []config.Action{{
					Type: "fixed-response",
					FixedResponse: config.FixedResponse{
						ContentType: "text/html",
						StatusCode:  200,
						MessageBody: "HELLO TEST!",
					},
				}},
			},
			{
				Conditions: []config.Condition{{
					PathPattern: &config.PathPatternCondition{
						Values: []string{"/auth/*"},
					},
				}},
				Actions: []config.Action{
					{
						Type: "authenticate-oidc",
						AuthenticateOIDC: config.AuthenticateOIDC{
							Scope:                    "",
							OnUnauthenticatedRequest: "deny",
						},
					},
					{
						Type: "forward",
						Forward: config.Forward{
							Protocol: "http",
							Host:     "localhost",
							Port:     8088,
						},
					},
				},
			},
		},
	}).Start()
}
