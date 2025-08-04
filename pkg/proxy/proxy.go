package proxy

import (
	"context"
	"errors"
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

type (
	forwardContextKey struct{}
	authContextKey    struct{}
)

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

		authData, ok := r.Context().Value(authContextKey{}).(authData)
		if ok {
			// TODO add other auth headers
			r.Header.Set("x-amzn-oidc-identity", authData.username)
		}

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
		actions := match(conf, r)
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
				shouldContinue, authData, err := handleAuthenticateOIDCAction(action.AuthenticateOIDC, w, r)
				if err != nil {
					writeErr(w, r, 500, "Internal server error")
					return
				}
				if !shouldContinue {
					return
				}
				if authData != nil {
					// overwrite r to include auth context for next action in stack
					r = r.WithContext(context.WithValue(r.Context(), authContextKey{}, *authData))
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

	code := http.StatusMovedPermanently
	if redirect.StatusCode == "HTTP_302" {
		code = http.StatusFound
	}

	http.Redirect(w, r, u.String(), code)
	return nil
}

func handleFixedResponseAction(fixedResponse config.FixedResponse, w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", fixedResponse.ContentType)
	w.WriteHeader(fixedResponse.StatusCode)
	_, _ = w.Write([]byte(fixedResponse.MessageBody))
}

func handleAuthenticateOIDCAction(authOIDC config.AuthenticateOIDC, w http.ResponseWriter, r *http.Request) (bool, *authData, error) {
	authData, err := authenticate(r)
	if err != nil {
		return false, nil, err
	}

	// if user is authenticated
	if authData != nil {
		return true, authData, nil
	}

	switch authOIDC.OnUnauthenticatedRequest {
	case config.OnUnauthenticatedRequestDeny:
		writeErr(w, r, 401, "Unauthorized")
		return false, nil, nil
	case config.OnUnauthenticatedRequestAllow:
		return true, nil, nil
	case config.OnUnauthenticatedRequestAuthenticate:
		// TODO send requested scopes
		http.Redirect(w, r, "/alb/auth", http.StatusFound)
		return false, nil, nil
	}

	return false, nil, errors.New("unexpected OnUnauthenticatedRequest in config")
}

func writeErr(w http.ResponseWriter, r *http.Request, code int, message string) {
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(code)
	_, _ = w.Write([]byte(message))
}

// TODO move to service
type authData struct {
	username string
}

// TODO move to service
func authenticate(r *http.Request) (*authData, error) {
	// return nil, nil
	return &authData{username: "ggg"}, nil
}
