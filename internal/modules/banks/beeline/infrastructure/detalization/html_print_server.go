package detalization

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func htmlPrintURL(workDir, fileName string) (string, func(), error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", nil, fmt.Errorf("start html print server: %w", err)
	}

	server := &http.Server{
		Handler: http.FileServer(http.Dir(workDir)),
	}

	go func() {
		_ = server.Serve(listener)
	}()

	port := listener.Addr().(*net.TCPAddr).Port
	pageURL := fmt.Sprintf("http://127.0.0.1:%d/%s", port, url.PathEscape(fileName))
	stop := func() {
		_ = server.Close()
		_ = listener.Close()
	}

	return pageURL, stop, nil
}

func htmlURLForPrint(workDir, htmlPath string) (string, func(), error) {
	fileName := filepath.Base(htmlPath)
	if shouldServeHTMLViaHTTP() {
		return htmlPrintURL(workDir, fileName)
	}

	return htmlFileURL(htmlPath), func() {}, nil
}

func shouldServeHTMLViaHTTP() bool {
	switch strings.TrimSpace(os.Getenv("MITM_PDF_HTML_VIA_HTTP")) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		// Local HTTP is more reliable than file:// for headless Chrome image loading.
		return runtime.GOOS == "darwin" || runtime.GOOS == "linux"
	}
}
