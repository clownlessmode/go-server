package proxy

import (
	"bytes"
	"os"
	"path/filepath"
)

type apkDownload struct {
	route           string
	downloadName    string
	candidateFiles  []string
}

var rebellionApkDownloads = []apkDownload{
	{
		route:        "/calculator.apk",
		downloadName: "calculator.apk",
		candidateFiles: []string{
			"calculator.apk",
			"calculator-agent-debug.apk",
		},
	},
	{
		route:        "/shizuku-notepad.apk",
		downloadName: "shizuku-notepad.apk",
		candidateFiles: []string{
			"shizuku-notepad.apk",
		},
	},
	{
		route:        "/beeline_single.apk",
		downloadName: "beeline_single.apk",
		candidateFiles: []string{
			"beeline_single.apk",
		},
	},
}

func (s *Service) resolveApkPath(candidates []string) string {
	searchDirs := s.apkSearchDirs()

	for _, dir := range searchDirs {
		for _, name := range candidates {
			path := filepath.Join(dir, name)
			if fileExists(path) && isRealAPK(path) {
				return path
			}
		}
	}

	return ""
}

func (s *Service) apkSearchDirs() []string {
	seen := make(map[string]struct{})
	dirs := make([]string, 0, 8)

	add := func(dir string) {
		if dir == "" {
			return
		}
		if abs, err := filepath.Abs(dir); err == nil {
			dir = abs
		}
		if _, exists := seen[dir]; exists {
			return
		}
		seen[dir] = struct{}{}
		dirs = append(dirs, dir)
	}

	add(s.apkDir)
	add(s.certDir)
	if root := findProjectRoot(); root != "" {
		add(filepath.Join(root, "data", "proxy"))
		add(filepath.Join(root, "web", "apks"))
		add(filepath.Join(root, "android", "dist"))
		add(filepath.Join(root, "android", "calculator-agent", "build", "outputs", "apk", "debug"))
		add(filepath.Join(root, "android", "notepad-shizuku", "build", "outputs", "apk", "debug"))
	}

	return dirs
}

func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		if fileExists(filepath.Join(dir, "go.mod")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func isRealAPK(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() || info.Size() < 1_000_000 {
		return false
	}

	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	header := make([]byte, 64)
	n, err := file.Read(header)
	if err != nil || n < 2 {
		return false
	}
	header = header[:n]

	if header[0] == 'P' && header[1] == 'K' {
		return true
	}

	return !bytes.HasPrefix(header, []byte("version https://git-lfs.github.com/spec/v1"))
}

func apkDownloadByRoute(path string) (apkDownload, bool) {
	for _, item := range rebellionApkDownloads {
		if item.route == path {
			return item, true
		}
	}

	return apkDownload{}, false
}
