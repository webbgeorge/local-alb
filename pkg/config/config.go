package config

const (
	ActionTypeForward          = "forward"
	ActionTypeRedirect         = "redirect"
	ActionTypeFixedResponse    = "fixed-response"
	ActionTypeAuthenticateOIDC = "authenticate-oidc"
)

// TODO validation
// TODO default values
type Config struct {
	Port           int
	DefaultActions []Action // only multiple if first is auth action
	Rules          []Rule
}

type Rule struct {
	Actions    []Action // only multiple if first is auth action
	Conditions []Condition
}

type Action struct {
	Type             string
	AuthenticateOIDC AuthenticateOIDC
	FixedResponse    FixedResponse
	Forward          Forward
	Redirect         Redirect
	// TODO implement AuthenticateCognito
}

type Condition struct {
	HostHeader        string
	HTTPHeader        string
	HTTPRequestMethod string
	PathPattern       string
	QueryString       string
	SourceIP          string
}

type AuthenticateOIDC struct {
	OnUnauthenticatedRequest string
	Scope                    string
}

type FixedResponse struct {
	ContentType string
	MessageBody string
	StatusCode  int
}

type Forward struct {
	Host     string // represents the target group
	Port     int
	Protocol string
}

type Redirect struct {
	StatusCode string
	Protocol   string
	Host       string
	Port       int
	Path       string
	Query      string
}
