package main

import (
	"net/http"
	"strconv"

	"github.com/aymerick/raymond"
)

// errorCatcher is a wrapper for http.ResponseWriter that
// captures 4xx and 5xx status codes and handles them in
// a custom manner
type errorCatcher struct {
	req          *http.Request
	res          http.ResponseWriter
	errorTpl     *raymond.Template
	notFoundTpl  *raymond.Template
	handledError bool
}

func (ec *errorCatcher) Header() http.Header {
	return ec.res.Header()
}

func (ec *errorCatcher) Write(buf []byte) (int, error) {
	// if we have already sent a response, pretend that this was successful
	if ec.handledError {
		return len(buf), nil
	}
	return ec.res.Write(buf)
}

func (ec *errorCatcher) WriteHeader(statusCode int) {
	if ec.handledError {
		return
	}
	if statusCode == 404 {
		ctx := map[string]string{
			"path": ec.req.URL.Path,
		}
		page, err := ec.notFoundTpl.Exec(ctx)
		// if we don't have a page to write, return before
		// we toggle the flag so we fall back to the original
		// error page
		if err != nil {
			return
		}
		ec.res.Header().Set("Content-Type", "text/html; charset=utf-8")
		ec.res.WriteHeader(statusCode)
		ec.res.Write([]byte(page))
		ec.handledError = true
		return
	}

	if statusCode >= 400 && statusCode < 600 {
		ctx := map[string]string{
			"code": strconv.Itoa(statusCode),
		}
		page, err := ec.errorTpl.Exec(ctx)
		// if we don't have a page to write, return before
		// we toggle the flag so we fall back to the original
		// error page
		if err != nil {
			return
		}
		ec.res.Header().Set("Content-Type", "text/html; charset=utf-8")
		ec.res.WriteHeader(statusCode)
		ec.res.Write([]byte(page))
		ec.handledError = true
		return
	}

	ec.res.WriteHeader(statusCode)
}
