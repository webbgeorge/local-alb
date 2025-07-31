package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/webbgeorge/local-alb/pkg/config"
)

type Proxy struct {
	server *http.Server
}

func NewProxy(conf config.Config) *Proxy {
	revProxy := &httputil.ReverseProxy{
		Director:     director(),
		ErrorHandler: errorHandler(),
	}
	return &Proxy{
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", conf.Port),
			Handler:      handler(conf, revProxy),
			ReadTimeout:  time.Second * 60, // TODO configurable
			WriteTimeout: time.Second * 60, // TODO configurable
		},
	}
}

func (p *Proxy) Start() {
	err := p.server.ListenAndServe()
	if err != nil {
		panic(err) // TODO better err handling
	}
}

type forwardContextKey struct{}

func director() func(r *http.Request) {
	return func(r *http.Request) {
		forward, ok := r.Context().Value(forwardContextKey{}).(config.Forward)
		if !ok {
			// this state shouldn't be possible to reach
			panic("attempted to forward request, but no forward config was provided")
		}

		// remove headers which shouldn't be set by incoming requests
		r.Header.Del("x-amzn-oidc-accesstoken")
		r.Header.Del("x-amzn-oidc-identity")
		r.Header.Del("x-amzn-oidc-data")
		r.Header.Del("X-Forwarded-For")

		// TODO get auth context if applicable and add headers

		r.URL.Scheme = forward.Protocol
		r.URL.Host = fmt.Sprintf("%s:%d", forward.Host, forward.Port)
	}
}

func errorHandler() func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		// TODO log error
		w.WriteHeader(502)
		_, _ = w.Write([]byte("Bad gateway"))
	}
}

// TODO logging
func handler(conf config.Config, revProxy *httputil.ReverseProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actions, err := match(conf, r)
		if err != nil {
			writeErr(w, r, 500, "Internal server error")
			return
		}

		if len(actions) == 0 {
			if len(conf.DefaultActions) == 0 {
				writeErr(w, r, 500, "Internal server error")
				return
			}
			actions = conf.DefaultActions
		}

		for _, action := range actions {
			switch action.Type {
			case config.ActionTypeForward:
				handleForwardAction(action.Forward, w, r, revProxy)
				return
			case config.ActionTypeRedirect:
				err := handleRedirectAction(action.Redirect, w, r)
				if err != nil {
					writeErr(w, r, 500, "Internal server error")
					return
				}
				return
			case config.ActionTypeFixedResponse:
				handleFixedResponseAction(action.FixedResponse, w, r)
				return
			case config.ActionTypeAuthenticateOIDC:
				shouldContinue, err := handleAuthenticateOIDCAction(action.AuthenticateOIDC, w, r)
				if err != nil {
					writeErr(w, r, 500, "Internal server error")
					return
				}
				if !shouldContinue {
					return
				}
			}
		}

		writeErr(w, r, 500, "Internal server error")
	}
}

func handleForwardAction(forward config.Forward, w http.ResponseWriter, r *http.Request, revProxy *httputil.ReverseProxy) {
	r = r.WithContext(context.WithValue(r.Context(), forwardContextKey{}, forward))
	revProxy.ServeHTTP(w, r)
}

func handleRedirectAction(redirect config.Redirect, w http.ResponseWriter, r *http.Request) error {
	u := &url.URL{}

	u.Scheme = redirect.Protocol // TODO handle replace of #{protocol}
	// TODO handle replace of #{host} and #{port}
	u.Host = fmt.Sprintf("%s:%d", redirect.Host, redirect.Port)
	u.Path = redirect.Path      // TODO handle replace of #{path}, #{host}, #{port}
	u.RawQuery = redirect.Query // TODO handle replace of #{query}, #{path}, #{host}, #{port}

	code := 301
	if redirect.StatusCode == "HTTP_302" {
		code = 302
	}

	http.Redirect(w, r, u.String(), code)
	return nil
}

func handleFixedResponseAction(fixedResponse config.FixedResponse, w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", fixedResponse.ContentType)
	w.WriteHeader(fixedResponse.StatusCode)
	_, _ = w.Write([]byte(fixedResponse.MessageBody))
}

func handleAuthenticateOIDCAction(authOIDC config.AuthenticateOIDC, w http.ResponseWriter, r *http.Request) (bool, error) {
	writeErr(w, r, 401, "Unauthorized") // TODO implement
	return false, nil                   // TODO
}

func writeErr(w http.ResponseWriter, r *http.Request, code int, message string) {
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(code)
	_, _ = w.Write([]byte(message))
}
