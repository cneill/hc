package hc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
)

type ResponseOpt func(response *http.Response) (*http.Response, error)

type ResponseOpts []ResponseOpt

func CopyRaw(writer io.Writer) ResponseOpt {
	return func(response *http.Response) (*http.Response, error) {
		response.Body = io.NopCloser(io.TeeReader(response.Body, writer))
		return response, nil
	}
}

func JSONResponse(holder any) ResponseOpt {
	return func(response *http.Response) (*http.Response, error) {
		// TODO: handle this explicitly elsewhere?
		defer response.Body.Close()

		decoder := json.NewDecoder(response.Body)
		if err := decoder.Decode(holder); err != nil {
			return nil, fmt.Errorf("JSONResponse: failed to decode: %w", err)
		}

		return response, nil
	}
}

func BodyRead(response *http.Response) (*http.Response, error) {
	_, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	return response, nil
}

var Status2XX = []int{
	http.StatusOK,
	http.StatusCreated,
	http.StatusAccepted,
	http.StatusNonAuthoritativeInfo,
	http.StatusNoContent,
	http.StatusResetContent,
	http.StatusPartialContent,
	http.StatusMultiStatus,
	http.StatusAlreadyReported,
	http.StatusIMUsed,
}

type StatusCodeError struct {
	StatusCode int
	Status     string
}

func (s StatusCodeError) Error() string {
	return fmt.Sprintf("status code: %d, message: %q", s.StatusCode, s.Status)
}

func AllowedStatusCodes(statusCodes ...int) ResponseOpt {
	return func(response *http.Response) (*http.Response, error) {
		if !slices.Contains(statusCodes, response.StatusCode) {
			return nil, StatusCodeError{
				StatusCode: response.StatusCode,
				Status:     response.Status,
			}
		}

		return response, nil
	}
}
