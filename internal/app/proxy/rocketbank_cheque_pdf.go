package proxy

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	rocketbankChequePDFPath = "/v1/reports/cheque-pdf"
	rocketbankChequePDFDir  = "data/reports/rocketbank/cheques"
	rocketbankChequePDFName = "rocket-reciept.pdf"
)

func (s *Service) rememberRocketbankHistoryTransaction(req *http.Request) {
	if req == nil ||
		req.Method != http.MethodGet ||
		!isRocketbankHost(req.Host) ||
		pathForLog(req) != rocketbankHistoryTransactionPath {
		return
	}

	transactionID := strings.TrimSpace(req.URL.Query().Get("transactionId"))
	if transactionID == "" {
		return
	}

	s.mu.Lock()
	s.lastRocketbankTransactionID = transactionID
	s.mu.Unlock()
}

func (s *Service) applyRocketbankChequePDFFallback(req *http.Request, res *http.Response) bool {
	if !isRocketbankMissingChequePDFRequest(req, res) {
		return false
	}

	transactionID := s.lastRocketbankHistoryTransaction()
	if transactionID == "" {
		proxyLog.Warnf("rocketbank cheque pdf fallback skipped: no previous transaction id")
		return false
	}

	path := filepath.Join(rocketbankChequePDFDir, filepath.Base(transactionID)+".pdf")
	body, err := os.ReadFile(path)
	if err != nil {
		proxyLog.Warnf("rocketbank cheque pdf fallback skipped: transactionId=%s err=%v", transactionID, err)
		return false
	}

	if res.Body != nil {
		if err := res.Body.Close(); err != nil {
			proxyLog.Warnf("rocketbank cheque pdf fallback response close failed: err=%v", err)
		}
	}

	res.StatusCode = http.StatusOK
	res.Status = "200 OK"
	res.Body = io.NopCloser(bytes.NewReader(body))
	res.ContentLength = int64(len(body))
	if res.Header == nil {
		res.Header = make(http.Header)
	}
	res.Header.Set("Content-Type", "application/pdf")
	res.Header.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, rocketbankChequePDFName))
	res.Header.Set("Content-Length", strconv.Itoa(len(body)))
	res.Header.Del("Content-Encoding")

	proxyLog.Infof("rocketbank cheque pdf fallback applied: transactionId=%s path=%s", transactionID, path)
	return true
}

func (s *Service) lastRocketbankHistoryTransaction() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastRocketbankTransactionID
}

func (s *Service) saveRocketbankChequePDF(req *http.Request, res *http.Response) bool {
	if !isRocketbankChequePDFRequest(req, res) {
		return false
	}

	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		proxyLog.Warnf("rocketbank cheque pdf read failed: err=%v", err)
		return true
	}
	if err := res.Body.Close(); err != nil {
		proxyLog.Warnf("rocketbank cheque pdf response close failed: err=%v", err)
	}
	res.Body = io.NopCloser(bytes.NewReader(rawBody))

	body, err := rocketbankChequePDFBody(rawBody, res.Header.Get("Content-Encoding"))
	if err != nil {
		proxyLog.Warnf("rocketbank cheque pdf decode failed: err=%v", err)
		return true
	}
	if len(body) == 0 {
		proxyLog.Warnf("rocketbank cheque pdf empty body")
		return true
	}

	path := filepath.Join(rocketbankChequePDFDir, rocketbankChequePDFFilename(req))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		proxyLog.Warnf("rocketbank cheque pdf mkdir failed: err=%v", err)
		return true
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		proxyLog.Warnf("rocketbank cheque pdf save failed: err=%v", err)
		return true
	}

	proxyLog.Infof("rocketbank cheque pdf saved: path=%s size=%d", path, len(body))
	return true
}

func isRocketbankChequePDFRequest(req *http.Request, res *http.Response) bool {
	return req.Method == http.MethodPost &&
		res.StatusCode == http.StatusOK &&
		isRocketbankHost(req.Host) &&
		pathForLog(req) == rocketbankChequePDFPath
}

func isRocketbankMissingChequePDFRequest(req *http.Request, res *http.Response) bool {
	return req.Method == http.MethodPost &&
		res.StatusCode == http.StatusBadRequest &&
		isRocketbankHost(req.Host) &&
		pathForLog(req) == rocketbankChequePDFPath
}

func rocketbankChequePDFBody(rawBody []byte, encoding string) ([]byte, error) {
	if strings.EqualFold(encoding, "gzip") {
		reader, err := gzip.NewReader(bytes.NewReader(rawBody))
		if err != nil {
			return nil, err
		}
		defer reader.Close()

		return io.ReadAll(reader)
	}

	if strings.TrimSpace(encoding) != "" {
		return nil, fmt.Errorf("unsupported content encoding: %s", encoding)
	}

	return rawBody, nil
}

func rocketbankChequePDFFilename(req *http.Request) string {
	base := "чек " + time.Now().Format("15-04")
	filename := base + ".pdf"
	for index := 2; ; index++ {
		if _, err := os.Stat(filepath.Join(rocketbankChequePDFDir, filename)); os.IsNotExist(err) {
			return filename
		}

		filename = fmt.Sprintf("%s %d.pdf", base, index)
	}
}
