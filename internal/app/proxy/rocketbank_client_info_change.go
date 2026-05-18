package proxy

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"project/internal/modules/banks/rocketbank/domain"
)

const rocketbankClientInfoPath = "/v1/clients/info"

func (s *Service) applyRocketbankClientInfoChangeScript(req *http.Request, res *http.Response) {
	if !isRocketbankClientInfoRequest(req, res) || s.rocketbankRepo == nil {
		return
	}

	config, err := s.rocketbankRepo.GetConfig(req.Context())
	if err != nil {
		proxyLog.Warnf("rocketbank client info change config read failed: err=%v", err)
		return
	}
	if !hasRocketbankClientInfo(config.ClientInfo) {
		return
	}

	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		proxyLog.Warnf("rocketbank client info change response read failed: err=%v", err)
		return
	}
	if err := res.Body.Close(); err != nil {
		proxyLog.Warnf("rocketbank client info change response close failed: err=%v", err)
	}

	changedBody, changed, err := rocketbankClientInfoChangedBody(rawBody, res.Header.Get("Content-Encoding"), config.ClientInfo)
	if err != nil {
		proxyLog.Warnf("rocketbank client info change failed: err=%v", err)
		res.Body = io.NopCloser(bytes.NewReader(rawBody))
		return
	}
	if !changed {
		res.Body = io.NopCloser(bytes.NewReader(rawBody))
		return
	}

	res.Body = io.NopCloser(bytes.NewReader(changedBody))
	res.ContentLength = int64(len(changedBody))
	res.Header.Set("Content-Length", strconv.Itoa(len(changedBody)))

	proxyLog.Infof("rocketbank client info change applied: route=%s", pathForLog(req))
}

func isRocketbankClientInfoRequest(req *http.Request, res *http.Response) bool {
	return req.Method == http.MethodGet &&
		res.StatusCode == http.StatusOK &&
		isRocketbankHost(req.Host) &&
		pathForLog(req) == rocketbankClientInfoPath
}

func hasRocketbankClientInfo(clientInfo domain.ClientInfo) bool {
	return clientInfo.FirstName != nil &&
		clientInfo.MiddleName != nil &&
		clientInfo.LastName != nil
}

func rocketbankClientInfoChangedBody(rawBody []byte, encoding string, clientInfo domain.ClientInfo) ([]byte, bool, error) {
	body := rawBody
	encoded := false

	if strings.EqualFold(encoding, "gzip") {
		reader, err := gzip.NewReader(bytes.NewReader(rawBody))
		if err != nil {
			return nil, false, err
		}
		defer reader.Close()

		body, err = io.ReadAll(reader)
		if err != nil {
			return nil, false, err
		}
		encoded = true
	} else if strings.TrimSpace(encoding) != "" {
		return nil, false, nil
	}

	var response map[string]any
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, false, err
	}

	firstName := formatRocketbankNamePart(*clientInfo.FirstName)
	middleName := formatRocketbankNamePart(*clientInfo.MiddleName)
	lastName := formatRocketbankNamePart(*clientInfo.LastName)

	response["appeal"] = firstName
	response["firstName"] = firstName
	response["fullName"] = rocketbankFullName(firstName, middleName, lastName)
	response["lastName"] = lastName
	response["middleName"] = middleName
	if clientInfo.PhoneNumber != nil {
		response["phone"] = strings.TrimSpace(*clientInfo.PhoneNumber)
	}

	changedBody, err := json.Marshal(response)
	if err != nil {
		return nil, false, err
	}

	if !encoded {
		return changedBody, true, nil
	}

	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	if _, err := writer.Write(changedBody); err != nil {
		_ = writer.Close()
		return nil, false, err
	}
	if err := writer.Close(); err != nil {
		return nil, false, err
	}

	return compressed.Bytes(), true, nil
}

func formatRocketbankNamePart(value string) string {
	parts := strings.Fields(strings.TrimSpace(value))
	for i, part := range parts {
		parts[i] = formatRocketbankHyphenatedNamePart(part)
	}

	return strings.Join(parts, " ")
}

func formatRocketbankHyphenatedNamePart(value string) string {
	parts := strings.Split(value, "-")
	for i, part := range parts {
		parts[i] = formatRocketbankSimpleNamePart(part)
	}

	return strings.Join(parts, "-")
}

func formatRocketbankSimpleNamePart(value string) string {
	runes := []rune(strings.ToLower(value))
	for i, char := range runes {
		if unicode.IsLetter(char) {
			runes[i] = unicode.ToUpper(char)
			break
		}
	}

	return string(runes)
}

func rocketbankFullName(firstName string, middleName string, lastName string) string {
	lastInitial := rocketbankNameInitial(lastName)
	if lastInitial == "" {
		return strings.TrimSpace(firstName + " " + middleName)
	}

	return strings.TrimSpace(lastInitial + ". " + firstName + " " + middleName)
}

func rocketbankNameInitial(value string) string {
	for _, char := range value {
		if unicode.IsLetter(char) {
			return string(unicode.ToUpper(char))
		}
	}

	return ""
}
