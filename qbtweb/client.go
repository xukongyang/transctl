package qbtweb

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/kenshaw/transctl/tctypes"
	"golang.org/x/net/publicsuffix"
)

// Client is a qbittorrent web client.
type Client struct {
	// cl is the underlying http client.
	cl *http.Client

	// userAgent is the user agent string sent to the rpc host.
	userAgent string

	// credentialFallback are the fallback credentials to try with.
	credentialFallback []string

	// url is the remote url host.
	url string

	// authenticated is the authentication toggle.
	authenticated bool

	sync.Mutex
}

// NewClient creates a new qBittorrent web client.
func NewClient(opts ...ClientOption) *Client {
	cl := &Client{
		cl:        new(http.Client),
		userAgent: "qbtweb/0.1",
	}
	for _, o := range opts {
		o(cl)
	}
	if cl.url == "" {
		WithHost("localhost:8080")(cl)
	}
	return cl
}

// authenticate
func (cl *Client) authenticate(ctx context.Context) error {
	cl.Lock()
	defer cl.Unlock()
	if cl.authenticated {
		return nil
	}

	var err error
	cl.cl.Jar, err = cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		return err
	}

	// build creds
	urlstr, c, err := cl.buildRequestURL("auth/login")
	if err != nil {
		return err
	}
	creds := cl.credentialFallback
	if c != nil {
		creds = c
	}
	username, password := "", ""
	if creds != nil {
		username, password = creds[0], creds[1]
	}
	if username == "" && password == "" {
		cl.authenticated = true
		return nil
	}

	// build request and execute
	var buf bytes.Buffer
	contentType, err := buildRequestBody(&buf, map[string]interface{}{
		"username": username,
		"password": password,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", urlstr, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", cl.userAgent)
	req.Header.Set("Content-Type", contentType)
	res, err := cl.cl.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return ErrUnauthorizedUser
	}

	cl.authenticated = true

	return nil
}

// buildRequestURL builds the request URL for the passed method, extracting the
// username/password in the original URL.
func (cl *Client) buildRequestURL(method string) (string, []string, error) {
	u, err := url.Parse(strings.TrimSuffix(cl.url, "/") + "/" + method)
	if err != nil {
		return "", nil, err
	}
	var creds []string
	if u.User != nil {
		pass, _ := u.User.Password()
		creds = append(creds, u.User.Username(), pass)
		u.User = nil
	}
	return u.String(), creds, nil
}

// buildRequestData builds the request url for method, encoding the rguments to
// the passed writer. Additionally extracts and returns any credentials in the
// client url.
func (cl *Client) buildRequestData(method string, arguments interface{}, w io.Writer) (string, string, []string, error) {
	urlstr, creds, err := cl.buildRequestURL(method)
	if err != nil {
		return "", "", nil, err
	}

	// build form data
	if x, ok := arguments.(interface {
		EncodeFormData(io.Writer) (string, error)
	}); ok {
		contentType, err := x.EncodeFormData(w)
		if err != nil {
			return "", "", nil, err
		}
		return urlstr, contentType, creds, nil
	}

	// build params and encode
	m, err := buildParamMap(arguments)
	if err != nil {
		return "", "", nil, err
	}
	contentType, err := buildRequestBody(w, m)
	if err != nil {
		return "", "", nil, err
	}
	return urlstr, contentType, creds, nil
}

// Do executes the qbittorrent web method, json marshaling the passed
// arguments and unmarshaling the response to v (if provided).
func (cl *Client) Do(ctx context.Context, method string, arguments, v interface{}) error {
	var err error
	if err = cl.authenticate(ctx); err != nil {
		return err
	}

	// build url, params, body
	var buf bytes.Buffer
	urlstr, contentType, _, err := cl.buildRequestData(method, arguments, &buf)
	if err != nil {
		return err
	}

	/*
		asFormData, params, err := buildParamMap(arguments)
		if err != nil {
			return err
		}
		contentType, err := cl.buildRequestBody(asFormData, params)
		if err != nil {
			return err
		}
	*/

	// build request and execute
	req, err := http.NewRequest("POST", urlstr, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", cl.userAgent)
	req.Header.Set("Content-Type", contentType)
	res, err := cl.cl.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	switch {
	case res.StatusCode == http.StatusForbidden:
		return ErrUnauthorizedUser
	case res.StatusCode == http.StatusNotFound:
		return ErrTorrentNotFound
	case res.StatusCode == http.StatusUnsupportedMediaType:
		return ErrTorrentFileInvalid
	case res.StatusCode != http.StatusOK:
		return ErrRequestFailed
	}

	if v == nil {
		return nil
	}

	// decode
	dec := json.NewDecoder(res.Body)
	dec.DisallowUnknownFields()
	dec.UseNumber()
	return dec.Decode(v)
}

// AuthLogout executes a auth logout request.
func (cl *Client) AuthLogout(ctx context.Context) error {
	return AuthLogout().Do(ctx, cl)
}

// TorrentsInfo executes a torrents info request.
func (cl *Client) TorrentsInfo(ctx context.Context) ([]Torrent, error) {
	return TorrentsInfo().Do(ctx, cl)
}

// ClientOption is a qBittorrent web client option.
type ClientOption = func(*Client)

// WithURL is a qBittorrent web client option to set the remote URL.
func WithURL(urlstr string) ClientOption {
	return func(cl *Client) {
		cl.url = urlstr
	}
}

// WithHost is a qBittorrent web client option to set the remote host. Remote
// URL will become 'http://<host>/transmission/rpc'.
func WithHost(host string) ClientOption {
	return WithURL("http://" + host + "/api/v2")
}

// WithClient is a qBittorrent web client option to set the underlying
// http.Client used.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(cl *Client) {
		cl.cl = httpClient
	}
}

// WithTimeout is a qBittorrent web client option to set the rpc host request
// tiemout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(cl *Client) {
		cl.cl.Timeout = timeout
	}
}

// WithUserAgent is a qBittorrent web client option to set the user agent sent
// to the rpc host.
func WithUserAgent(userAgent string) ClientOption {
	return func(cl *Client) {
		cl.userAgent = userAgent
	}
}

// WithCredentialFallback is a qBittorrent web client option to set the
// credential fallback to send to the rpc host.
func WithCredentialFallback(user, pass string) ClientOption {
	return func(cl *Client) {
		cl.credentialFallback = []string{user, pass}
	}
}

// WithLogf is a qBittorrent web client option to set a logging handler for
// HTTP requests and responses.
func WithLogf(logf func(string, ...interface{})) ClientOption {
	return func(cl *Client) {
		cl.cl.Transport = tctypes.NewHTTPLogf(cl.cl.Transport, logf)
	}
}
