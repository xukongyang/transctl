package delrpc

import (
	"bytes"
	"compress/zlib"
	"context"
	"io"
	"net"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// DefaultTimeout is the default client timeout.
	DefaultTimeout = 10 * time.Second

	// DefaultUserAgent is the default client user agent.
	DefaultUserAgent = "delrpc/0.1"
)

// Client is a deluge rpc client.
type Client struct {
	// timeout is the dial timeout.
	timeout time.Duration

	// userAgent is the user agent string sent to the rpc host.
	userAgent string

	// credentialFallback are the fallback credentials to try with.
	credentialFallback []string

	// url is the remote url host.
	url string

	// conn is the net connection.
	conn net.Conn

	// id is the request id.
	id int64

	// authenticated indicates the client has already authenticated.
	authenticated bool

	// reqf is the logging function used to send requests.
	reqf func(string, ...interface{})

	// resf is the logging function used to send responses.
	resf func(string, ...interface{})

	sync.Mutex
}

// NewClient creates a new deluge rpc client.
func NewClient(opts ...ClientOption) *Client {
	cl := &Client{
		timeout:   DefaultTimeout,
		userAgent: DefaultUserAgent,
	}
	for _, o := range opts {
		o(cl)
	}
	if cl.url == "" {
		WithHost("localhost:8080")(cl)
	}
	return cl
}

// open opens a connection to the deluge rpc host.
func (cl *Client) open(ctx context.Context) error {
	u, err := url.Parse(cl.url)
	if err != nil {
		return err
	}
	d := net.Dialer{
		Timeout: cl.timeout,
	}
	cl.conn, err = d.DialContext(ctx, "tcp", u.Hostname()+":"+u.Port())
	return err
}

// Close closes the connection.
func (cl *Client) Close() error {
	cl.Lock()
	defer cl.Unlock()

	if cl.conn != nil {
		if err := cl.conn.Close(); err != nil {
			return err
		}
		cl.conn = nil
	}

	return nil
}

// do executes a request and reads the response.
func (cl *Client) do(ctx context.Context, method string, req, res interface{}) error {
	reqID := atomic.AddInt64(&cl.id, 1)

	// encode
	var err error
	var reqBuf bytes.Buffer
	if err = encode(&reqBuf, reqID, method, req); err != nil {
		return err
	}
	if cl.reqf != nil {
		cl.reqf("%s", string(reqBuf.Bytes()))
	}

	// write
	w := zlib.NewWriter(cl.conn)
	if _, err = w.Write(reqBuf.Bytes()); err != nil {
		return err
	}
	if err = w.Flush(); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}

	// read
	r, err := zlib.NewReader(cl.conn)
	if err != nil {
		return err
	}
	var resBuf bytes.Buffer
	if _, err = io.Copy(&resBuf, r); err != nil {
		return err
	}

	// decode
	if cl.resf != nil {
		cl.resf("%s", string(resBuf.Bytes()))
	}
	resID, err := decode(resBuf.Bytes(), res)
	if err != nil {
		return err
	}
	if reqID != resID {
		return ErrMismatchedRequestAndResponseIDs
	}
	return nil
}

// authenticate
func (cl *Client) authenticate(ctx context.Context) error {
	if cl.authenticated {
		return nil
	}

	// determine creds
	creds := cl.credentialFallback
	u, err := url.Parse(cl.url)
	if err != nil {
		return err
	}
	if u.User != nil {
		pass, _ := u.User.Password()
		creds = []string{u.User.Username(), pass}
	}

	if creds[0] == "" {
		cl.authenticated = true
		return nil
	}

	req := struct {
		Username string
		Password string
		Params   map[string]interface{}
	}{
		Username: creds[0],
		Password: creds[1],
		Params: map[string]interface{}{
			"client_version": "2.0.3",
		},
	}
	if err := cl.do(ctx, "daemon.login", req, nil); err != nil {
		return err
	}

	cl.authenticated = true
	return nil
}

// Do executes the qbittorrent web method, json marshaling the passed
// arguments and unmarshaling the response to v (if provided).
func (cl *Client) Do(ctx context.Context, method string, arguments, v interface{}) error {
	cl.Lock()
	defer cl.Unlock()

	var err error

	// open and authenticate
	if err = cl.open(ctx); err != nil {
		return err
	}
	if err = cl.authenticate(ctx); err != nil {
		return err
	}
	return nil
}

// ClientOption is a deluge rpc client option.
type ClientOption = func(*Client)

// WithURL is a deluge rpc client option to set the remote URL.
func WithURL(urlstr string) ClientOption {
	return func(cl *Client) {
		cl.url = urlstr
	}
}

// WithHost is a deluge rpc client option to set the remote host. Remote
// URL will become 'http://<host>/transmission/rpc'.
func WithHost(host string) ClientOption {
	return WithURL("http://" + host + "/api/v2")
}

// WithTimeout is a deluge rpc client option to set the rpc host request
// tiemout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(cl *Client) {
		cl.timeout = timeout
	}
}

// WithUserAgent is a deluge rpc client option to set the user agent sent
// to the rpc host.
func WithUserAgent(userAgent string) ClientOption {
	return func(cl *Client) {
		cl.userAgent = userAgent
	}
}

// WithCredentialFallback is a deluge rpc client option to set the
// credential fallback to send to the rpc host.
func WithCredentialFallback(user, pass string) ClientOption {
	return func(cl *Client) {
		cl.credentialFallback = []string{user, pass}
	}
}

// WithLogf is a deluge rpc client option to set logging handlers HTTP
// request and response bodies.
func WithLogf(reqf, resf func(string, ...interface{})) ClientOption {
	return func(cl *Client) {
		cl.reqf, cl.resf = reqf, resf
	}
}
