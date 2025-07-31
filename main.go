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
					PathPattern: "/test",
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
		},
	}).Start()
}
