package metrics

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Middleware struct {
	Handler http.Handler
	Metrics *Metrics
}

func NewMiddleware(m *Metrics) (*Middleware, error) {
	if m == nil {
		var err error
		m, err = GetObject(nil, nil)
		if err != nil {
			return nil, err
		}
	}
	return &Middleware{nil, m}, nil
}

func (self *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	self.Handler.ServeHTTP(w, r)

	// Method + adjusted URL to key base
	// e.g., GET /flavors/31963730-1630-4f30-a4cf-cea6ef9c19a5 => GET.flavors-X
	working := ""
	var found bool
	chunks := strings.Split(r.URL.Path, "/")
	chunks = chunks[1:]
	for i, chunk := range chunks {
		fmt.Println(i)
		fmt.Println(chunk)
		found = false
		if i == 0 {
			found = true
		} else {
			for _, allowed := range self.Metrics.config.UrlToKey.AllowedChunks {
				if chunk == allowed {
					found = true
					chunk = "-" + chunk
				}
			}
		}
		if found == false {
			chunk = "-X"
		}
		r1 := regexp.MustCompile("[^a-zA-Z0-9_.\\-]")
		chunk = r1.ReplaceAllString(chunk, "-")
		r2 := regexp.MustCompile("-+")
		chunk = r2.ReplaceAllString(chunk, "-")
		working += chunk
	}
	working = r.Method + "." + working

	self.Metrics.MeasureSince(makeKey(working, "time"), now)
	self.Metrics.IncrCounter(makeKey(working, "count"), 1)
}

func makeKey(base string, suffix string) []string {
	key := make([]string, 0)
	key = append(key, base+"."+suffix)
	return key
}
