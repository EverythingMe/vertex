package vertex

import (
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dvirsky/go-pylog/logging"

	"code.google.com/p/go-uuid/uuid"
	"golang.org/x/text/language"
)

// Request wraps the standard http request object with higher level contextual data
type Request struct {
	*http.Request
	StartTime time.Time
	Deadline  time.Time
	Locale    string
	UserAgent string
	RemoteIP  string
	Location  struct{ Lat, Long float64 }
	RequestId string
	Callback  string
	Secure    bool

	attributes map[string]interface{}
}

func (r *Request) SetAttribute(key string, val interface{}) {
	r.attributes[key] = val
}

func (r *Request) Attribute(key string) (interface{}, bool) {

	v, found := r.attributes[key]
	return v, found
}

const DefaultLocale = "en-US"

const HeaderGeoPosition = "X-LatLong"

// parse the locale based on Accept-Language header. If no header found or the values are invalid,
// we fall back to en-US
func (r *Request) parseLocale() {

	tags, _, err := language.ParseAcceptLanguage(r.Header.Get("Accept-Language"))
	if err != nil {
		logging.Warning("Could not parse accept lang header: %s", err)
		return
	}

	if len(tags) > 0 {
		logging.Debug("Locale for request: %s", tags[0])
		r.Locale = tags[0].String()
	}
}

// parse the client address, based on http headers or the actual ip
func (r *Request) parseAddr() {

	// the default is the originating ip. but we try to find better options because this is almost
	// never the right IP
	if parts := strings.Split(r.Request.RemoteAddr, ":"); len(parts) == 2 {
		r.RemoteIP = parts[0]
	}
	// If we have a forwarded-for header, take the address from there
	if xff := strings.Trim(r.Header.Get("X-Forwarded-For"), ","); len(xff) > 0 {
		addrs := strings.Split(xff, ",")
		lastFwd := addrs[len(addrs)-1]
		if ip := net.ParseIP(lastFwd); ip != nil {
			logging.Debug("Setting IP based on XFF header to %s", ip)
			r.RemoteIP = ip.String()
		}

	}
	if xri := r.Header.Get("X-Real-Ip"); len(xri) > 0 {
		if ip := net.ParseIP(xri); ip != nil {
			logging.Debug("Setting IP based on XRI header to %s", ip)
			r.RemoteIP = ip.String()
		}
	}

	logging.Debug("Request ip: %s", r.RemoteIP)

}

// Parse our location header
func (r *Request) parseLocation() {

	latlong := r.Header.Get(HeaderGeoPosition)
	if latlong != "" {

		parts := strings.Split(latlong, ",")
		if len(parts) == 2 {

			if lat, err := strconv.ParseFloat(parts[0], 64); err == nil {
				if long, err := strconv.ParseFloat(parts[1], 64); err == nil {
					r.Location.Lat = lat
					r.Location.Long = long
				}
			}

		}
	}

}

// Detect if the request is secure or not, based on either TLS info or http headers/url
func (r *Request) parseSecure() {

	if r.TLS != nil {
		r.Secure = true
		return
	}

	if u, err := url.ParseRequestURI(r.RequestURI); err == nil {

		if u.Scheme == "https" {
			r.Secure = true
			return
		}
	}

	xfp := r.Header.Get("X-Forwarded-Proto")
	if xfp == "" {
		xfp = r.Header.Get("X-Scheme")
	}

	if xfp == "https" {
		r.Secure = true
	}
}

// wrap a new http request with a vertex request
func newRequest(r *http.Request) *Request {
	req := &Request{
		Request:    r,
		StartTime:  time.Now(),
		Locale:     DefaultLocale,
		UserAgent:  r.UserAgent(),
		RequestId:  uuid.New(),
		Callback:   r.FormValue(CallbackParam),
		attributes: make(map[string]interface{}),
	}

	req.parseLocale()
	req.parseAddr()
	req.parseLocation()
	req.parseSecure()

	return req
}
