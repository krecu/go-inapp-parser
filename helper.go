package go_inapp_parser

import (
	"fmt"
	"time"

	"github.com/bitly/go-simplejson"

	"github.com/valyala/fasthttp"
)

const MaxRedirect int = 10

var ErrRedirectCount = fmt.Errorf("max redirect")

func newJSON(data []byte) (j *simplejson.Json, err error) {
	j, err = simplejson.NewJson(data)
	if err != nil {
		return nil, err
	}
	return j, nil
}

// doFollowRedirectsTimeout - request with support 301/302
func doFollowRedirectsTimeout(req *fasthttp.Request, res *fasthttp.Response, timeout time.Duration) error {

	tc := AcquireTimer(timeout)
	defer ReleaseTimer(tc)

	ch := make(chan error)
	go func() {
		ch <- doFollowRedirects(req, res)
	}()

	select {
	case err := <-ch:
		return err
	case <-tc.C:
		res.SetStatusCode(fasthttp.StatusRequestTimeout)
		return fasthttp.ErrTimeout
	}
}

func doFollowRedirects(req *fasthttp.Request, res *fasthttp.Response) error {

	var uri = string(req.RequestURI())
	var redirectCounter int = 0

	for redirectCounter <= MaxRedirect {

		req.SetRequestURI(uri)

		// run request
		if err := fasthttp.Do(req, res); err != nil {
			return err
		}

		// check complete status
		if !(res.Header.StatusCode() >= fasthttp.StatusMovedPermanently &&
			res.Header.StatusCode() <= fasthttp.StatusPermanentRedirect) {
			return nil
		}

		// check next redirect header
		if location := res.Header.Peek(fasthttp.HeaderLocation); location != nil {
			uri = string(location)
			redirectCounter++
		} else {
			break
		}
	}

	res.SetStatusCode(fasthttp.StatusRequestTimeout)
	return ErrRedirectCount
}
