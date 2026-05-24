package detalization

import (
	"os"
	"strings"
)

const (
	firstPageTemplateV1 = "templates/first-page.html"
	firstPageTemplateV2 = "templates/first-page-v2.html"

	envFirstPageType = "FIRST_PAGE_TYPE"
)

var firstPageTypeOverride string

func SetFirstPageType(version string) {
	firstPageTypeOverride = strings.TrimSpace(version)
}

func firstPageTemplate() string {
	switch strings.TrimSpace(firstPageTypeOverride) {
	case "2":
		return firstPageTemplateV2
	case "1":
		return firstPageTemplateV1
	}

	switch strings.TrimSpace(firstPageTypeEnv()) {
	case "2":
		return firstPageTemplateV2
	default:
		return firstPageTemplateV1
	}
}

func firstPageTypeEnv() string {
	for _, key := range []string{envFirstPageType, "FIRSTPAGETYPE"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}

	return ""
}
