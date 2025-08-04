package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/auth/aa", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = fmt.Fprintf(w, `
<div>
			username: %s
</div>
			`, r.Header.Get("x-amzn-oidc-identity"))
	})
	http.ListenAndServe(":8088", nil)
}
