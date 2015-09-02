package oauth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/dvirsky/go-pylog/logging"

	"gitlab.doit9.com/server/vertex"

	"golang.org/x/oauth2"
)

type TokenProtocol interface {
	EncodeToken(interface{}) (string, error)
	DecodeToken(string) (interface{}, error)
}
type UserValidator interface {
	TokenProtocol
	Login(token *oauth2.Token) (interface{}, error)
}
type OAuthMiddleware struct {
	conf          *oauth2.Config
	jwtKey        []byte
	userValidator UserValidator
}

// OAuth2 config. This is copied from the oauth2 library so it can be parsed from yaml with added tags
type Config struct {
	// ClientID is the application's ID.
	ClientID string `yaml:"client_id"`
	// ClientSecret is the application's secret.
	ClientSecret string `yaml:"client_secret"`

	// The provider's authentication url
	AuthURL string `yaml:"auth_url"`
	// The providers token fetching url
	TokenURL string `yaml:"token_url"`

	// RedirectURL is the URL to redirect users going through
	// the OAuth flow, after the resource owner's URLs.
	RedirectURL string `yaml:"redirect_url"`

	// Scope specifies optional requested permissions.
	Scopes []string `yaml:"scopes"`
}

func NewOAuthMiddleware(config *Config, validator UserValidator) *OAuthMiddleware {

	return &OAuthMiddleware{
		userValidator: validator,
		conf: &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RedirectURL:  config.RedirectURL,
			Scopes:       config.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  config.AuthURL,
				TokenURL: config.TokenURL,
			},
		},
	}
}

const (
	tokenKey  = "oauth...token"
	loginPath = "/login"
	AttrUser  = "oauth_user"
	nextUrl   = "next_url"
)

func (o *OAuthMiddleware) getToken(r *vertex.Request) (interface{}, error) {

	if cookie, err := r.Cookie(tokenKey); err == nil {

		user, err := o.userValidator.DecodeToken(cookie.Value)
		if err != nil {
			return nil, err
		}

		return user, nil

	}
	return "", errors.New("Could not get cookie")

}

type JWTAuthenticator struct {
	key []byte
}

func NewJWTAuthenticator(key string) *JWTAuthenticator {
	return &JWTAuthenticator{
		key: []byte(key),
	}
}

func (j *JWTAuthenticator) EncodeToken(data interface{}) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims["data"] = data

	sstr, err := token.SignedString(j.key)
	if err != nil {
		logging.Error("Error signing token: %s", err)

	}
	return sstr, err
}

func (j *JWTAuthenticator) Login(token *oauth2.Token) (interface{}, error) {
	return token.AccessToken, nil
}
func (j *JWTAuthenticator) DecodeToken(data string) (interface{}, error) {
	token, err := jwt.Parse(data, func(token *jwt.Token) (interface{}, error) {
		return j.key, nil
	})

	if err == nil && token.Valid {
		return token.Claims["data"].(string), nil

	} else {
		return "", logging.Errorf("Invalid token '%s' (%#v)! %s", data, token, err)
	}
}

func (o *OAuthMiddleware) setCookie(w http.ResponseWriter, data string, domain string) {

	cookie := &http.Cookie{
		Name:    tokenKey,
		Value:   data,
		Path:    "/",
		Domain:  domain,
		Expires: time.Now().Add(time.Hour * 24),
	}
	http.SetCookie(w, cookie)
}

func (o *OAuthMiddleware) redirect(w http.ResponseWriter, r *vertex.Request) {

	//save the current url for laterz
	http.SetCookie(w, &http.Cookie{
		Name:  nextUrl,
		Value: r.RequestURI,
		Path:  "/",
	})

	url := o.conf.AuthCodeURL("mystate", oauth2.AccessTypeOnline)
	http.Redirect(w, r.Request, url, 302)

}

func (o *OAuthMiddleware) LoginHandler() vertex.Route {

	handler := func(w http.ResponseWriter, r *vertex.Request) (interface{}, error) {
		code := r.FormValue("code")
		logging.Info("Got code: %s", code)

		tok, err := o.conf.Exchange(oauth2.NoContext, code)
		if err != nil {
			return nil, vertex.UnauthorizedError("Could not log you in: %s", err)
		}

		user, err := o.userValidator.Login(tok)
		if err != nil {
			return nil, vertex.UnauthorizedError("Could not validate user for login: %s", err)
		}

		enc, err := o.userValidator.EncodeToken(user)
		if err != nil {
			return nil, vertex.UnauthorizedError("Could not validate encode user token: %s", err)
		}

		o.setCookie(w, enc, r.Host)

		if cook, err := r.Cookie(nextUrl); err == nil && cook != nil && cook.Value != "" {
			logging.Info("Found nextUrl from before auth denied. Redirecting to %s", cook.Value)
			http.Redirect(w, r.Request, cook.Value, http.StatusTemporaryRedirect)
			return nil, vertex.Hijacked
		}

		return "Success Logging In", nil
	}
	return vertex.Route{
		Path:        loginPath,
		Description: "OAuth Login",
		Handler:     vertex.HandlerFunc(handler),
		Methods:     vertex.GET,
	}

}

func (o *OAuthMiddleware) Handle(w http.ResponseWriter, r *vertex.Request, next vertex.HandlerFunc) (interface{}, error) {

	if strings.HasSuffix(r.URL.Path, loginPath) {
		return next(w, r)
	}
	user, err := o.getToken(r)
	if err != nil {
		o.redirect(w, r)
		return nil, vertex.Hijacked

	}

	logging.Info("Request authenticated. Continuing!")
	r.SetAttribute(AttrUser, user)

	return next(w, r)
}
