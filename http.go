package checkup

import (
	"net/http"
	"net/url"
	"time"
)

func doHTTPCall(api API) (int, error) {
	u, err := url.Parse(api.Endpoint)
	if err != nil {
		debugLog(err)
		return 0, err
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		debugLog(err)
		return 0, err
	}
	req.Close = true

	if api.Timeout == 0 {
		api.Timeout = 1000
	}
	hc := &http.Client{
		Timeout: time.Duration(api.Timeout) * time.Millisecond,
	}

	resp, err := hc.Do(req)
	if err != nil {
		debugLog(err)
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}
