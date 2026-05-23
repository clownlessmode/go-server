package proxy

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"strings"
)

func decodeBeelineResponseBody(rawBody []byte, encoding string) ([]byte, bool, error) {
	if strings.EqualFold(encoding, "gzip") {
		reader, err := gzip.NewReader(bytes.NewReader(rawBody))
		if err != nil {
			return nil, false, err
		}
		defer reader.Close()

		body, err := io.ReadAll(reader)
		if err != nil {
			return nil, false, err
		}

		return body, true, nil
	}
	if strings.TrimSpace(encoding) != "" {
		return nil, false, nil
	}

	return rawBody, false, nil
}

func encodeBeelineResponseBody(body []byte, gzipEncoded bool) ([]byte, error) {
	if !gzipEncoded {
		return body, nil
	}

	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	if _, err := writer.Write(body); err != nil {
		_ = writer.Close()
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	return compressed.Bytes(), nil
}

func readBeelineJSONResponse(rawBody []byte, encoding string) (map[string]any, []byte, bool, error) {
	body, encoded, err := decodeBeelineResponseBody(rawBody, encoding)
	if err != nil {
		return nil, rawBody, false, err
	}
	if body == nil {
		return nil, rawBody, encoded, nil
	}

	var response map[string]any
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, rawBody, encoded, err
	}

	return response, rawBody, encoded, nil
}

func writeBeelineJSONResponse(response map[string]any, rawBody []byte, encoded bool) ([]byte, bool, error) {
	changedBody, err := json.Marshal(response)
	if err != nil {
		return rawBody, false, err
	}

	out, err := encodeBeelineResponseBody(changedBody, encoded)
	if err != nil {
		return rawBody, false, err
	}

	return out, true, nil
}
