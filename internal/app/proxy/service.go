package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	adguardlog "github.com/AdguardTeam/golibs/log"
	"github.com/AdguardTeam/gomitmproxy"
	"github.com/AdguardTeam/gomitmproxy/mitm"
	"github.com/AdguardTeam/gomitmproxy/proxyutil"

	"project/internal/app/config"
	"project/internal/app/logger"
	beelinedomain "project/internal/modules/banks/beeline/domain"
	rocketbankdomain "project/internal/modules/banks/rocketbank/domain"
	smssend "project/internal/modules/sms/usecase/send"
	"project/internal/modules/banks/beeline/usecase/recordpaymentflow"
)

var proxyLog = logger.New("proxy")

type Service struct {
	cfg            config.ProxyConfig
	smsCfg         config.SMSConfig
	rocketbankCfg  config.RocketbankConfig
	proxy          *gomitmproxy.Proxy
	certDir        string
	apkDir         string
	rocketbankRepo rocketbankdomain.Repository
	beelineRepo    beelinedomain.Repository
	smsSend        *smssend.UseCase
	recordPaymentFlow *recordpaymentflow.UseCase

	mu                          sync.Mutex
	beelineLogMu                sync.Mutex
	beelinePaymentMu            sync.Mutex
	beelinePaymentContext       beelinePaymentSnapshot
	beelineSimMu                sync.Mutex
	activeBeelineSimNumber      string
	beelineProductCTNs          []string
	beelineSessionMu            sync.Mutex
	beelineSessionHeaderMap     map[string]string
	beelineRefreshHTTPClient    *http.Client
	beelinePendingCardTemplate  string
	beelineDetalizationTaskMu   sync.Mutex
	beelineDetalizationTasks    map[string]beelineDetalizationTaskParams
	lastBeelineDetalizationTask beelineDetalizationTaskParams
	fakeOdid                    string
	lastRocketbankTransactionID string
}

func NewService(
	cfg config.ProxyConfig,
	smsCfg config.SMSConfig,
	rocketbankCfg config.RocketbankConfig,
	rocketbankRepo rocketbankdomain.Repository,
	beelineRepo beelinedomain.Repository,
	smsSend *smssend.UseCase,
	recordPaymentFlow *recordpaymentflow.UseCase,
) (*Service, error) {
	adguardlog.SetOutput(io.Discard)

	ca, err := loadOrCreateCA(cfg.CertDir)
	if err != nil {
		return nil, err
	}

	mitmConfig, err := mitm.NewConfig(ca.cert, ca.key, nil)
	if err != nil {
		return nil, fmt.Errorf("create mitm config: %w", err)
	}
	mitmConfig.SetValidity(7 * 24 * time.Hour)
	mitmConfig.SetOrganization("Rebellion")

	addr, err := resolveTCPAddr(cfg.Address)
	if err != nil {
		return nil, err
	}

	service := &Service{
		cfg:            cfg,
		smsCfg:         smsCfg,
		rocketbankCfg:  rocketbankCfg,
		certDir:        cfg.CertDir,
		apkDir:         cfg.ApkDir,
		rocketbankRepo: rocketbankRepo,
		beelineRepo:       beelineRepo,
		smsSend:           smsSend,
		recordPaymentFlow: recordPaymentFlow,
		fakeOdid:       newRandomUUID(),
	}

	proxyLog.Infof("odid_spoof: using fake ODID: %s", service.fakeOdid)

	service.proxy = gomitmproxy.NewProxy(gomitmproxy.Config{
		ListenAddr: addr,
		MITMConfig: mitmConfig,
		OnRequest:  service.handleRequest,
		OnResponse: service.handleResponse,
		OnError:    service.handleError,
	})

	return service, nil
}

func (s *Service) Start() error {
	if err := s.proxy.Start(); err != nil {
		return err
	}

	proxyLog.Successf("started on %s, install page: http://%s", s.cfg.Address, s.cfg.Host)
	if s.cfg.RocketbankLogs {
		proxyLog.Infof("rocketbank response logging enabled: %s", rocketbankLogFile)
	}
	if s.cfg.BeelineLogs {
		proxyLog.Infof("proxy response logging enabled: %s", beelineLogFile)
	}
	if s.smsCfg.Enabled {
		proxyLog.Infof("sms agent queue enabled")
	}
	return nil
}

func (s *Service) Close() {
	if s == nil || s.proxy == nil {
		return
	}

	s.proxy.Close()
}

func (s *Service) handleRequest(session *gomitmproxy.Session) (*http.Request, *http.Response) {
	req := session.Request()
	s.rememberRocketbankHistoryTransaction(req)

	if s.isMagicHost(req.Host) {
		if req.Method == http.MethodConnect {
			return nil, badRequest(req, "Open http://"+s.cfg.Host+" instead of HTTPS.")
		}

		proxyLog.Infof("install page request: path=%s", req.URL.Path)
		return nil, s.handleMagicHost(req)
	}

	s.logShalltryRequest(req)
	s.captureBeelineSession(req)
	s.captureBeelineRequestForLog(session, req)
	s.captureBeelinePaymentRequest(req)
	if res := s.maybeSpoofShalltryOdid(req); res != nil {
		return nil, res
	}
	if res := s.maybeSpoofBeelineCatalogTransaction(req); res != nil {
		return nil, res
	}

	return nil, nil
}

func (s *Service) handleResponse(session *gomitmproxy.Session) *http.Response {
	req := session.Request()
	res := session.Response()
	if res == nil {
		return nil
	}
	if !session.Ctx().IsMITM() || req.Method == http.MethodConnect {
		return nil
	}

	s.logShalltryResponse(req, res)
	s.captureBeelinePaymentResponse(req, res)
	s.captureBeelineActiveSim(req, res)
	s.rememberBeelineDetalizationTask(session, req, res)

	s.applyRocketbankBalanceChangeScript(req, res)
	s.applyBeelineDetalizationChangeScript(req, res)
	beelineDetalizationReportHandled := s.applyBeelineDetalizationReportScript(req, res)
	s.applyBeelineBalanceChangeScript(req, res)
	s.applyRocketbankCardInfoChangeScript(req, res)
	s.applyRocketbankClientInfoChangeScript(req, res)
	s.applyRocketbankHistoryChangeScript(req, res)
	s.applyRocketbankHistoryTransactionChangeScript(req, res)
	rocketbankChequeSaved := s.saveRocketbankChequePDF(req, res)
	rocketbankChequeHandled := rocketbankChequeSaved || s.applyRocketbankChequePDFFallback(req, res)
	if s.cfg.RocketbankLogs && isRocketbankHost(req.Host) && !rocketbankChequeHandled {
		s.writeRocketbankResponseLog(req, res)
	}
	if s.cfg.BeelineLogs && !s.isMagicHost(req.Host) && !isBeelineLogExcludedHost(req.Host) && !beelineDetalizationReportHandled {
		s.writeBeelineResponseLog(session, req, res)
	}
	if res.StatusCode >= 400 {
		return nil
	}

	proxyLog.Successf("%s %s%s -> %d", req.Method, req.Host, req.URL.Path, res.StatusCode)

	return nil
}

func (s *Service) handleError(session *gomitmproxy.Session, err error) {
}

func (s *Service) handleMagicHost(req *http.Request) *http.Response {
	switch req.URL.Path {
	case "", "/":
		proxyLog.Infof("serving install page")
		return htmlResponse(req, installPage(s.cfg.Host))
	case "/android.cer", "/cert.cer":
		proxyLog.Infof("serving android certificate")
		return fileResponse(req, filepath.Join(s.certDir, certCERFile), "rebellion-ca-cert.cer", "application/x-x509-ca-cert")
	case "/ios.pem", "/ca-cert.pem", "/cert.pem":
		proxyLog.Infof("serving ios certificate")
		return fileResponse(req, filepath.Join(s.certDir, certPEMFile), "rebellion-ca-cert.pem", "application/x-x509-ca-cert")
	default:
		if apk, ok := apkDownloadByRoute(req.URL.Path); ok {
			path := s.resolveApkPath(apk.candidateFiles)
			if path == "" {
				proxyLog.Warnf("apk not found: route=%s", req.URL.Path)
				return notFound(req)
			}
			proxyLog.Infof("serving apk: route=%s path=%s", req.URL.Path, path)
			return fileResponse(req, path, apk.downloadName, "application/vnd.android.package-archive")
		}
		proxyLog.Warnf("install page not found: path=%s", req.URL.Path)
		return notFound(req)
	}
}

func htmlResponse(req *http.Request, body string) *http.Response {
	res := proxyutil.NewResponse(http.StatusOK, bytes.NewBufferString(body), req)
	res.Header.Set("Content-Type", "text/html; charset=utf-8")
	return res
}

func fileResponse(req *http.Request, path string, filename string, contentType string) *http.Response {
	body, err := os.ReadFile(path)
	if err != nil {
		return notFound(req)
	}

	res := proxyutil.NewResponse(http.StatusOK, bytes.NewReader(body), req)
	res.Header.Set("Content-Type", contentType)
	res.Header.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	return res
}

func notFound(req *http.Request) *http.Response {
	res := proxyutil.NewResponse(http.StatusNotFound, bytes.NewBufferString("not found"), req)
	res.Header.Set("Content-Type", "text/plain; charset=utf-8")
	return res
}

func badRequest(req *http.Request, message string) *http.Response {
	res := proxyutil.NewResponse(http.StatusBadRequest, bytes.NewBufferString(message), req)
	res.Header.Set("Content-Type", "text/plain; charset=utf-8")
	return res
}

func resolveTCPAddr(address string) (*net.TCPAddr, error) {
	if strings.HasPrefix(address, ":") {
		address = "0.0.0.0" + address
	}

	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("resolve proxy address: %w", err)
	}

	return addr, nil
}

func (s *Service) isMagicHost(requestHost string) bool {
	return isMagicHost(requestHost, s.cfg.Host) || isMagicHost(requestHost, "rebellion.com")
}

func isMagicHost(requestHost string, magicHost string) bool {
	host := requestHost
	if splitHost, _, err := net.SplitHostPort(requestHost); err == nil {
		host = splitHost
	}

	return strings.EqualFold(host, magicHost)
}
