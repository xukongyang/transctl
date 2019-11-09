package transrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"sync/atomic"
)

// Client is a transmission rpc client.
type Client struct {
	http.Client

	// url is the remote url host.
	url string

	// retries is the number of times to retry on a error, such as a 409
	// (missing csrf id) error.
	retries int

	// tag is an incrementing number used for each rpc request and response.
	tag int64

	// csrf is the CSRF session id.
	csrf string

	sync.RWMutex
}

// NewClient issues a new transmission rpc client.
func NewClient(opts ...ClientOption) *Client {
	cl := &Client{
		url:     "http://transmission:transmission@localhost:9091/transmission/rpc/",
		retries: 3,
	}
	for _, o := range opts {
		o(cl)
	}
	return cl
}

// Do executes the transmission rpc method, json marshaling the passed
// arguments and unmarshaling the response to v (if provided).
func (cl *Client) Do(ctx context.Context, method string, arguments, v interface{}) error {
	// encode rpc body
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(map[string]interface{}{
		"method":    method,
		"arguments": arguments,
		"tag":       atomic.AddInt64(&cl.tag, 1),
	}); err != nil {
		return err
	}

	// execute, retrying as per rpc spec
	var res *http.Response
	for i := 0; res == nil || res.StatusCode == 409 || i > cl.retries; i++ {
		// create http request
		req, err := http.NewRequest("POST", cl.url, bytes.NewReader(body.Bytes()))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		cl.RLock()
		if cl.csrf != "" {
			req.Header.Set(csrfHeader, cl.csrf)
		}
		cl.RUnlock()

		// execute
		res, err = cl.Client.Do(req.WithContext(ctx))
		if err != nil {
			return err
		}
		defer res.Body.Close()
		cl.Lock()
		if csrf := res.Header.Get(csrfHeader); csrf != "" {
			cl.csrf = csrf
		}
		cl.Unlock()
	}

	// check status
	if res == nil || res.StatusCode != 200 {
		return ErrRequestFailed
	}

	/*
		buf, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		log.Printf(">>> RESPONSE: %s", string(buf))
	*/

	// setup result
	result := struct {
		Result    string      `json:"result,omitempty"`
		Arguments interface{} `json:"arguments,omitempty"`
		Tag       int64       `json:"tag,omitempty"`
	}{
		Arguments: v,
	}

	// decode
	dec := json.NewDecoder(res.Body)
	dec.DisallowUnknownFields()
	dec.UseNumber()
	if err := dec.Decode(&result); err != nil {
		return err
	}

	// check success
	if result.Result != "success" {
		return ErrRequestFailed
	}

	return nil
}

// TorrentStart issues a torrent start request for the specified ids.
func (cl *Client) TorrentStart(ctx context.Context, ids ...interface{}) error {
	return TorrentStart(ids...).Do(ctx, cl)
}

// TorrentStartNow issues a torrent start now request for the specified ids.
func (cl *Client) TorrentStartNow(ctx context.Context, ids ...interface{}) error {
	return TorrentStartNow(ids...).Do(ctx, cl)
}

// TorrentStop issues a torrent stop request for the specified ids.
func (cl *Client) TorrentStop(ctx context.Context, ids ...interface{}) error {
	return TorrentStop(ids...).Do(ctx, cl)
}

// TorrentVerify issues a torrent verify request for the specified ids.
func (cl *Client) TorrentVerify(ctx context.Context, ids ...interface{}) error {
	return TorrentVerify(ids...).Do(ctx, cl)
}

// TorrentReannounce issues a torrent reannounce request for the specified ids.
func (cl *Client) TorrentReannounce(ctx context.Context, ids ...interface{}) error {
	return TorrentReannounce(ids...).Do(ctx, cl)
}

// TorrentSet issues a torrent start request for the specified ids.
func (cl *Client) TorrentSet(ctx context.Context, req *TorrentSetRequest) error {
	return req.Do(ctx, cl)
}

// TorrentGet issues a torrent get request for the specified ids.
func (cl *Client) TorrentGet(ctx context.Context, ids ...interface{}) (*TorrentGetResponse, error) {
	return TorrentGet(ids...).Do(ctx, cl)
}

// TorrentAdd issues a torrent add request for the specified ids.
func (cl *Client) TorrentAdd(ctx context.Context, req *TorrentAddRequest) (*TorrentAddResponse, error) {
	return req.Do(ctx, cl)
}

// TorrentRemove issues a torrent remove request for the specified ids.
func (cl *Client) TorrentRemove(ctx context.Context, deleteLocalData bool, ids ...interface{}) error {
	return TorrentRemove(deleteLocalData, ids...).Do(ctx, cl)
}

// TorrentSetLocation issues a torrent set location request for the specified ids.
func (cl *Client) TorrentSetLocation(ctx context.Context, location string, move bool, ids ...interface{}) error {
	return TorrentSetLocation(location, move, ids...).Do(ctx, cl)
}

// TorrentRenamePath issues a torrent start request for the specified ids.
func (cl *Client) TorrentRenamePath(ctx context.Context, path, name string, ids ...interface{}) error {
	return TorrentRenamePath(path, name, ids...).Do(ctx, cl)
}

// SessionSet issues a session set request.
func (cl *Client) SessionSet(ctx context.Context, req *SessionSetRequest) error {
	return req.Do(ctx, cl)
}

// SessionGet issues a session get request.
func (cl *Client) SessionGet(ctx context.Context) (*SessionGetResponse, error) {
	return SessionGet().Do(ctx, cl)
}

// SessionStats issues a session stats request.
func (cl *Client) SessionStats(ctx context.Context) (*SessionStatsResponse, error) {
	return SessionStats().Do(ctx, cl)
}

// BlocklistUpdate issues a blocklist update request.
func (cl *Client) BlocklistUpdate(ctx context.Context) (int64, error) {
	return BlocklistUpdate().Do(ctx, cl)
}

// PortTest issues a port test request.
func (cl *Client) PortTest(ctx context.Context) (bool, error) {
	return PortTest().Do(ctx, cl)
}

// SessionClose issues a session close request.
func (cl *Client) SessionClose(ctx context.Context) error {
	return SessionClose().Do(ctx, cl)
}

// SessionShutdown issues a session close request.
//
// Alias for SessionClose.
func (cl *Client) SessionShutdown(ctx context.Context) error {
	return SessionClose().Do(ctx, cl)
}

// QueueMoveTop creates a queue move top request for the specified ids.
func (cl *Client) QueueMoveTop(ctx context.Context, ids ...interface{}) error {
	return QueueMoveTop(ids...).Do(ctx, cl)
}

// QueueMoveUp creates a queue move up request for the specified ids.
func (cl *Client) QueueMoveUp(ctx context.Context, ids ...interface{}) error {
	return QueueMoveUp(ids...).Do(ctx, cl)
}

// QueueMoveDown creates a queue move down request for the specified ids.
func (cl *Client) QueueMoveDown(ctx context.Context, ids ...interface{}) error {
	return QueueMoveDown(ids...).Do(ctx, cl)
}

// QueueMoveBottom creates a queue move bottom request for the specified ids.
func (cl *Client) QueueMoveBottom(ctx context.Context, ids ...interface{}) error {
	return QueueMoveBottom(ids...).Do(ctx, cl)
}

// FreeSpace issues a free space request.
func (cl *Client) FreeSpace(ctx context.Context, path string) (int64, error) {
	return FreeSpace(path).Do(ctx, cl)
}

// ClientOption is a transmission rpc client option.
type ClientOption = func(*Client)

// WithURL is a transmission rpc client option to set the remote URL.
func WithURL(urlstr string) ClientOption {
	return func(cl *Client) {
		cl.url = urlstr
	}
}

// WithHost is a transmission rpc client option to set the remote host. Remote
// URL will become 'http://<host>/transmission/rpc'.
func WithHost(host string) ClientOption {
	return WithURL("http://" + host + "/transmission/rpc/")
}

// WithCSRF is a transmission rpc client option to set the CSRF token used.
func WithCSRF(csrf string) ClientOption {
	return func(cl *Client) {
		cl.csrf = csrf
	}
}

// WithRetries is a transmission rpc client option to set the number of retry
// attempts when a failure is encountered.
func WithRetries(retries int) ClientOption {
	return func(cl *Client) {
		cl.retries = retries
	}
}
