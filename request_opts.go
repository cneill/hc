package hc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type RequestOpt func(*http.Request) (*http.Request, error)

type RequestOpts []RequestOpt

func Combine(opts ...RequestOpt) RequestOpt {
	return func(request *http.Request) (*http.Request, error) {
		var err error

		for _, opt := range opts {
			request, err = opt(request)
			if err != nil {
				return nil, fmt.Errorf("opt error: %w", err)
			}
		}

		return request, nil
	}
}

func NewBaseRequestFactory(opts ...RequestOpt) func() *http.Request {
	return func() *http.Request {
		req, err := NewRequest(opts...)
		if err != nil {
			panic(err)
		}

		return req
	}
}

func Get(request *http.Request) (*http.Request, error) {
	request.Method = http.MethodGet
	return request, nil
}

func Head(request *http.Request) (*http.Request, error) {
	request.Method = http.MethodHead
	return request, nil
}

func Post(request *http.Request) (*http.Request, error) {
	request.Method = http.MethodPost
	return request, nil
}

func Put(request *http.Request) (*http.Request, error) {
	request.Method = http.MethodPut
	return request, nil
}

func Patch(request *http.Request) (*http.Request, error) {
	request.Method = http.MethodPatch
	return request, nil
}

func Delete(request *http.Request) (*http.Request, error) {
	request.Method = http.MethodDelete
	return request, nil
}

func Connect(request *http.Request) (*http.Request, error) {
	request.Method = http.MethodConnect
	return request, nil
}

func Options(request *http.Request) (*http.Request, error) {
	request.Method = http.MethodOptions
	return request, nil
}

func Trace(request *http.Request) (*http.Request, error) {
	request.Method = http.MethodTrace
	return request, nil
}

func UserAgent(userAgent string) RequestOpt {
	return func(request *http.Request) (*http.Request, error) {
		request.Header.Add("User-Agent", userAgent)
		return request, nil
	}
}

func Context(ctx context.Context) RequestOpt {
	return func(request *http.Request) (*http.Request, error) {
		return request.WithContext(ctx), nil
	}
}

func JSONRequest(body any) RequestOpt {
	return func(request *http.Request) (*http.Request, error) {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("WithJSONBody: failed to marshal: %w", err)
		}

		request.Body = io.NopCloser(bytes.NewReader(jsonBytes))
		request.Header.Set("Content-Type", "application/json")

		return request, nil
	}
}

func BaseURL(base string) RequestOpt {
	return func(request *http.Request) (*http.Request, error) {
		parsed, err := url.Parse(base)
		if err != nil {
			return nil, fmt.Errorf("WithBaseURL: failed to parse URL: %w", err)
		}

		request.URL = parsed

		return request, nil
	}
}

func Path(path string) RequestOpt {
	return func(request *http.Request) (*http.Request, error) {
		if request.URL == nil {
			return nil, fmt.Errorf("WithPath: requires base URL")
		}

		request.URL.Path = path

		return request, nil
	}
}

func Query(key string, value any) RequestOpt {
	return func(request *http.Request) (*http.Request, error) {
		values := url.Values{}

		if request.URL.RawQuery != "" {
			existingValues, err := url.ParseQuery(request.URL.RawQuery)
			if err != nil {
				return nil, fmt.Errorf("QueryVar: %w", err)
			}

			values = existingValues
		}

		stringValue, err := valueToString(value)
		if err != nil {
			return nil, fmt.Errorf("QueryVar: %w", err)
		}

		values.Add(key, stringValue)

		request.URL.RawQuery = values.Encode()

		return request, nil
	}
}

func valueToString(value any) (string, error) { //nolint:cyclop,funlen //Stupid type assertions...
	var (
		result    string
		converted bool
	)

	switch typed := value.(type) {
	case string:
		result = typed
		converted = true
	case bool:
		result = strconv.FormatBool(typed)
		converted = true
	case int:
		result = strconv.FormatInt(int64(typed), 10)
		converted = true
	case int8:
		result = strconv.FormatInt(int64(typed), 10)
		converted = true
	case int16:
		result = strconv.FormatInt(int64(typed), 10)
		converted = true
	case int32:
		result = strconv.FormatInt(int64(typed), 10)
		converted = true
	case int64:
		result = strconv.FormatInt(typed, 10)
		converted = true
	case uint:
		result = strconv.FormatUint(uint64(typed), 10)
		converted = true
	case uint8:
		result = strconv.FormatUint(uint64(typed), 10)
		converted = true
	case uint16:
		result = strconv.FormatUint(uint64(typed), 10)
		converted = true
	case uint32:
		result = strconv.FormatUint(uint64(typed), 10)
		converted = true
	case uint64:
		result = strconv.FormatUint(typed, 10)
		converted = true
	case float32:
		result = strconv.FormatFloat(float64(typed), 'g', -1, 32)
		converted = true
	case float64:
		result = strconv.FormatFloat(typed, 'g', -1, 64)
		converted = true
	}

	if stringer, ok := value.(fmt.Stringer); ok {
		result = stringer.String()
		converted = true
	}

	if converted {
		return result, nil
	}

	return fancyToString(value)
}

func fancyToString(value any) (string, error) {
	if stringer, ok := value.(fmt.Stringer); ok {
		return stringer.String(), nil
	}

	renderedVal := fmt.Sprintf("%+v", value)
	if len(renderedVal) > 25 {
		renderedVal = renderedVal[:25] + "..."
	}

	return "", fmt.Errorf("failed to convert value of type \"%T\" (%s) to string", value, renderedVal)
}
