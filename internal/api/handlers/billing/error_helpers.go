package billing

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

// writeErrorFromErr logs err server-side with context and writes a generic
// client-visible message via the caller-supplied writeError. Use in place of
// `h.writeError(w, status, code, err.Error())` to prevent leaking internal
// err text (SQL errors, file paths, library names, etc.) to clients.
//
// `writeFn` should be the handler's existing writeError method, e.g.
// `h.writeError` for *CostAllocationHandler, `h.writeError` for
// *ExternalBillingHandler, etc. The code argument is passed through; the
// message is replaced with a generic, status-appropriate one.
func writeErrorFromErr(
	r *http.Request,
	w http.ResponseWriter,
	status int,
	code, contextMsg string,
	err error,
	writeFn func(http.ResponseWriter, int, string, string),
) {
	if err != nil {
		fields := logrus.Fields{
			"status":  status,
			"code":    code,
			"context": contextMsg,
			"method":  "",
			"path":    "",
		}
		if r != nil {
			fields["method"] = r.Method
			if r.URL != nil {
				fields["path"] = r.URL.Path
			}
		}
		entry := logrus.WithError(err).WithFields(fields)
		if status >= 500 {
			entry.Error("billing handler error")
		} else {
			entry.Info("billing handler client error")
		}
	}
	writeFn(w, status, code, sanitizedBillingMessage(status, code))
}

func sanitizedBillingMessage(status int, code string) string {
	if status >= 500 {
		return "Internal server error"
	}
	switch code {
	case "Invalid Request":
		return "Invalid request"
	case "Unauthorized":
		return "Unauthorized"
	case "Not Found":
		return "Not found"
	case "Forbidden":
		return "Forbidden"
	case "Conflict":
		return "Conflict"
	}
	return "Request failed"
}

// writeErrorFromMessage wraps the package-level `writeError` (used by
// state_usage_handler.go and similar files that have a free function rather
// than a method) so that err is logged server-side and the user-facing
// message is a static, generic one.
//
// `userMsg` is a hand-written, client-safe description (e.g. "Invalid date
// range"). It is always used; err is never forwarded to the client.
func writeErrorFromMessage(r *http.Request, w http.ResponseWriter, status int, userMsg, contextMsg string, err error) {
	if err != nil {
		fields := logrus.Fields{
			"status":  status,
			"context": contextMsg,
			"method":  "",
			"path":    "",
		}
		if r != nil {
			fields["method"] = r.Method
			if r.URL != nil {
				fields["path"] = r.URL.Path
			}
		}
		entry := logrus.WithError(err).WithFields(fields)
		if status >= 500 {
			entry.Error("billing handler error")
		} else {
			entry.Info("billing handler client error")
		}
	}
	writeError(w, status, userMsg)
}
