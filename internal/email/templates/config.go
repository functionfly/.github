package templates

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type EmailConfig struct {
	CompanyName    string
	CompanyDomain  string
	DashboardURL   string
	StatusPageURL  string
	SupportURL     string
	PrivacyURL     string
	TermsURL       string
}

var (
	cfg      = &EmailConfig{}
	cfgMutex sync.RWMutex
)

func SetEmailConfig(c EmailConfig) {
	cfgMutex.Lock()
	defer cfgMutex.Unlock()
	cfg = &c
}

func getConfig() EmailConfig {
	cfgMutex.RLock()
	defer cfgMutex.RUnlock()
	if cfg.CompanyDomain == "" {
		return EmailConfig{
			CompanyName:   "FunctionFly",
			CompanyDomain:  "functionfly.com",
			DashboardURL:   "https://dashboard.functionfly.com",
			StatusPageURL:  "https://status.functionfly.com",
			SupportURL:     "https://functionfly.com/support",
			PrivacyURL:     "https://functionfly.com/privacy",
			TermsURL:       "https://functionfly.com/terms",
		}
	}
	return *cfg
}

func InitEmailConfigFromEnv() {
	companyName := os.Getenv("EMAIL_COMPANY_NAME")
	companyDomain := os.Getenv("EMAIL_COMPANY_DOMAIN")
	dashboardURL := os.Getenv("EMAIL_DASHBOARD_URL")
	statusPageURL := os.Getenv("EMAIL_STATUS_PAGE_URL")
	supportURL := os.Getenv("EMAIL_SUPPORT_URL")
	privacyURL := os.Getenv("EMAIL_PRIVACY_URL")
	termsURL := os.Getenv("EMAIL_TERMS_URL")

	if companyDomain == "" && dashboardURL == "" {
		return
	}

	emailCfg := EmailConfig{}
	if companyName != "" {
		emailCfg.CompanyName = companyName
	} else {
		emailCfg.CompanyName = "FunctionFly"
	}

	if companyDomain != "" {
		emailCfg.CompanyDomain = companyDomain
	} else {
		emailCfg.CompanyDomain = "functionfly.com"
	}

	if dashboardURL != "" {
		emailCfg.DashboardURL = dashboardURL
	} else {
		emailCfg.DashboardURL = fmt.Sprintf("https://dashboard.%s", emailCfg.CompanyDomain)
	}

	if statusPageURL != "" {
		emailCfg.StatusPageURL = statusPageURL
	} else {
		emailCfg.StatusPageURL = fmt.Sprintf("https://status.%s", emailCfg.CompanyDomain)
	}

	if supportURL != "" {
		emailCfg.SupportURL = supportURL
	} else {
		emailCfg.SupportURL = fmt.Sprintf("https://%s/support", emailCfg.CompanyDomain)
	}

	if privacyURL != "" {
		emailCfg.PrivacyURL = privacyURL
	} else {
		emailCfg.PrivacyURL = fmt.Sprintf("https://%s/privacy", emailCfg.CompanyDomain)
	}

	if termsURL != "" {
		emailCfg.TermsURL = termsURL
	} else {
		emailCfg.TermsURL = fmt.Sprintf("https://%s/terms", emailCfg.CompanyDomain)
	}

	SetEmailConfig(emailCfg)
}

func BuildDashboardURL(paths ...string) string {
	cfg := getConfig()
	url := cfg.DashboardURL
	for _, p := range paths {
		url = fmt.Sprintf("%s/%s", strings.TrimSuffix(url, "/"), strings.TrimPrefix(p, "/"))
	}
	return url
}

func BuildFunctionDashboardURL(functionID string, subpath ...string) string {
	base := fmt.Sprintf("%s/functions/%s", getConfig().DashboardURL, functionID)
	if len(subpath) > 0 {
		return fmt.Sprintf("%s/%s", strings.TrimSuffix(base, "/"), strings.TrimPrefix(subpath[0], "/"))
	}
	return base
}

func TransactionalEmailCopyrightHTML() string {
	cfg := getConfig()
	return fmt.Sprintf(`© %d %s. All rights reserved.`, time.Now().Year(), cfg.CompanyName)
}

func TransactionalEmailCopyrightPlain() string {
	cfg := getConfig()
	return fmt.Sprintf("© %d %s. All rights reserved.", time.Now().Year(), cfg.CompanyName)
}

func TransactionalEmailCopyrightOrangeHTML() string {
	cfg := getConfig()
	return fmt.Sprintf(`© %d %s. All rights reserved.<br>
<a href="%s" style="color:#f97316;text-decoration:none;">Privacy Policy</a> ·
<a href="%s" style="color:#f97316;text-decoration:none;">Terms of Service</a>`, time.Now().Year(), cfg.CompanyName, cfg.PrivacyURL, cfg.TermsURL)
}

func TransactionalEmailCopyrightOrangePlain() string {
	cfg := getConfig()
	return fmt.Sprintf("© %d %s. All rights reserved.\n%s | %s", time.Now().Year(), cfg.CompanyName, cfg.PrivacyURL, cfg.TermsURL)
}
