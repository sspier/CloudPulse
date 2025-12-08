package probe

import (
	"context"
	"net/http"
	"time"

	"github.com/sspier/cloudpulse/internal/model"
)

// check performs a single probe of the target url
func Check(ctx context.Context, t model.Target) model.Result {
	startTime := time.Now()

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, t.URL, nil)
	// if creating request fails, it's a hard failure
	if err != nil {
		return model.Result{
			TargetID:   t.ID,
			Status:     "down",
			HTTPStatus: 0,
			Timestamp:  startTime.Unix(),
		}
	}

	// perform the probe with a timeout of 5 seconds
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// response from the probe
	httpResponse, err := client.Do(httpRequest)
	status := "down"
	httpStatus := 0

	if err == nil {
		httpStatus = httpResponse.StatusCode
		// if the http status is between 200 and 400, the target is up
		if httpResponse.StatusCode >= 200 && httpResponse.StatusCode < 400 {
			status = "up"
		}
		httpResponse.Body.Close()
	}

	return model.Result{
		TargetID:   t.ID,
		Status:     status,
		HTTPStatus: httpStatus,
		Timestamp:  startTime.Unix(),
	}
}
