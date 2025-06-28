package hc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
