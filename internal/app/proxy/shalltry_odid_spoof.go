package proxy

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/AdguardTeam/gomitmproxy/proxyutil"
)

const (
	shalltryOdidHost = "ire-oneid.shalltry.com"
	shalltryOdidPath = "/one/v1/odid"
	shalltryOdidTTL  = 2592000
)

type shalltryOdidResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Odid string `json:"odid"`
	Time int    `json:"time"`
}

func newRandomUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf(
		"%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16],
	)
}

func isShalltryHost(requestHost string) bool {
	return strings.Contains(strings.ToLower(hostForLog(requestHost)), "shalltry.com")
}

func isShalltryOdidRequest(req *http.Request) bool {
	host := strings.ToLower(hostForLog(req.Host))
	if !strings.Contains(host, shalltryOdidHost) {
		return false
	}

	return strings.Contains(pathForLog(req), shalltryOdidPath)
}

func (s *Service) logShalltryRequest(req *http.Request) {
	if !isShalltryHost(req.Host) {
		return
	}

	host := hostForLog(req.Host)
	path := pathForLog(req)
	proxyLog.Infof("shalltry --> %s https://%s%s", req.Method, host, path)

	if req.Body == nil {
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	if len(body) == 0 {
		return
	}

	preview := string(body)
	if len(preview) > 500 {
		preview = preview[:500]
	}
	proxyLog.Infof("shalltry request body: %s", preview)
}

func (s *Service) logShalltryResponse(req *http.Request, res *http.Response) {
	if !isShalltryHost(req.Host) {
		return
	}

	proxyLog.Infof("shalltry <-- %d %s", res.StatusCode, pathForLog(req))

	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}
	if err := res.Body.Close(); err != nil {
		proxyLog.Warnf("shalltry response body close failed: route=%s err=%v", pathForLog(req), err)
	}
	res.Body = io.NopCloser(bytes.NewReader(rawBody))
	if len(rawBody) == 0 {
		return
	}

	responseBody := responseBodyForLog(rawBody, res.Header.Get("Content-Encoding"))
	preview := responseBody
	if len(preview) > 500 {
		preview = preview[:500]
	}
	proxyLog.Infof("shalltry response: %s", preview)
}

func (s *Service) maybeSpoofShalltryOdid(req *http.Request) *http.Response {
	if !isShalltryOdidRequest(req) {
		return nil
	}

	proxyLog.Infof("odid_spoof: intercepting ODID request, returning fake success")

	body, err := json.Marshal(shalltryOdidResponse{
		Code: 0,
		Msg:  "OK",
		Odid: s.fakeOdid,
		Time: shalltryOdidTTL,
	})
	if err != nil {
		proxyLog.Warnf("odid_spoof: marshal response failed: err=%v", err)
		return nil
	}

	return jsonResponse(req, body)
}

func jsonResponse(req *http.Request, body []byte) *http.Response {
	res := proxyutil.NewResponse(http.StatusOK, bytes.NewReader(body), req)
	res.Header.Set("Content-Type", "application/json; charset=utf-8")
	return res
}
