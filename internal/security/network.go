package security

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

// Network security scanning functions

func (sas *SecurityAuditService) scanNetworkSecurity(ctx context.Context, target string, config ScanConfig) ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	// Check SSL/TLS configuration
	if tlsVulns, err := sas.checkSSLConfiguration(target); err == nil {
		vulnerabilities = append(vulnerabilities, tlsVulns...)
	}

	// Check for open ports (basic port scanning simulation)
	if portVulns, err := sas.checkOpenPorts(target, config.IncludePorts); err == nil {
		vulnerabilities = append(vulnerabilities, portVulns...)
	}

	// Check firewall configuration
	if fwVulns, err := sas.checkFirewallConfiguration(); err == nil {
		vulnerabilities = append(vulnerabilities, fwVulns...)
	}

	return vulnerabilities, nil
}

func (sas *SecurityAuditService) checkSSLConfiguration(target string) ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	// Security audit tool intentionally skips verification to inspect certificate properties.
	// This is standard practice for TLS security scanners - it never makes real connections
	// to serve application traffic.
	conn, err := tls.Dial("tcp", target+":443", &tls.Config{
		InsecureSkipVerify: true, // Intentional: security audit needs to inspect cert validity
	})
	if err != nil {
		return vulnerabilities, err
	}
	defer conn.Close()

	cert := conn.ConnectionState().PeerCertificates[0]

	// Check certificate expiry
	daysUntilExpiry := int(cert.NotAfter.Sub(time.Now()).Hours() / 24)
	if daysUntilExpiry < 30 {
		vulnerabilities = append(vulnerabilities, Vulnerability{
			ID:          generateVulnID(),
			Title:       "SSL Certificate Expiring Soon",
			Description: fmt.Sprintf("SSL certificate expires in %d days", daysUntilExpiry),
			Severity:    "high",
			Category:    "crypto",
			Component:   "ssl",
			Location:    target,
			Status:      "open",
			Remediation: "Renew SSL certificate before expiry",
			Discovered:  time.Now(),
			Updated:     time.Now(),
		})
	}

	// Check for weak cipher suites
	state := conn.ConnectionState()
	if state.CipherSuite < tls.TLS_RSA_WITH_AES_128_CBC_SHA { // Basic check for very weak ciphers
		vulnerabilities = append(vulnerabilities, Vulnerability{
			ID:          generateVulnID(),
			Title:       "Weak SSL Cipher Suite",
			Description: "Server is using a weak SSL/TLS cipher suite",
			Severity:    "medium",
			Category:    "crypto",
			Component:   "ssl",
			Location:    target,
			Status:      "open",
			Remediation: "Configure server to use strong cipher suites (AES-256-GCM, ChaCha20-Poly1305)",
			Discovered:  time.Now(),
			Updated:     time.Now(),
		})
	}

	return vulnerabilities, nil
}

func (sas *SecurityAuditService) checkOpenPorts(target string, ports []int) ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	// Common vulnerable ports to check
	vulnerablePorts := map[int]string{
		21:  "FTP",
		23:  "Telnet",
		25:  "SMTP",
		53:  "DNS",
		110: "POP3",
		143: "IMAP",
		445: "SMB",
		3389: "RDP",
	}

	for port, service := range vulnerablePorts {
		if sas.isPortOpen(target, port) {
			vulnerabilities = append(vulnerabilities, Vulnerability{
				ID:          generateVulnID(),
				Title:       fmt.Sprintf("Potentially Vulnerable Port Open: %d (%s)", port, service),
				Description: fmt.Sprintf("Port %d (%s) is open and may be vulnerable to attacks", port, service),
				Severity:    "medium",
				Category:    "network",
				Component:   service,
				Location:    fmt.Sprintf("%s:%d", target, port),
				Status:      "open",
				Remediation: fmt.Sprintf("Close port %d or restrict access with firewall rules", port),
				Discovered:  time.Now(),
				Updated:     time.Now(),
			})
		}
	}

	return vulnerabilities, nil
}

func (sas *SecurityAuditService) checkFirewallConfiguration() ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	// Critical ports that should be protected by firewall rules
	criticalPorts := map[int]string{
		22:   "SSH",
		3306: "MySQL",
		5432: "PostgreSQL",
		6379: "Redis",
		27017: "MongoDB",
		9200: "Elasticsearch",
	}

	// Check for unprotected critical services
	for port, service := range criticalPorts {
		if sas.isPortOpen("localhost", port) {
			vulnerabilities = append(vulnerabilities, Vulnerability{
				ID:          generateVulnID(),
				Title:       fmt.Sprintf("Critical Service Exposed: %s (Port %d)", service, port),
				Description: fmt.Sprintf("%s service on port %d is accessible without firewall protection", service, port),
				Severity:    "high",
				Category:    "network",
				Component:   "firewall",
				Location:    fmt.Sprintf("localhost:%d", port),
				Status:      "open",
				Remediation: fmt.Sprintf("Configure firewall to restrict access to %s on port %d to authorized networks only", service, port),
				Discovered:  time.Now(),
				Updated:     time.Now(),
			})
		}
	}

	// Check for common firewall bypass techniques
	if sas.isPortOpen("localhost", 80) && sas.isPortOpen("localhost", 443) {
		// Check if HTTP redirects to HTTPS (basic web server protection)
		if vuln := sas.checkHTTPRedirect(); vuln != nil {
			vulnerabilities = append(vulnerabilities, *vuln)
		}
	}

	// Check for exposed administrative interfaces
	adminPorts := map[int]string{
		8080: "Admin Interface",
		8443: "Admin SSL",
		9000: "Admin Panel",
		10000: "Webmin",
	}

	for port, service := range adminPorts {
		if sas.isPortOpen("localhost", port) {
			vulnerabilities = append(vulnerabilities, Vulnerability{
				ID:          generateVulnID(),
				Title:       fmt.Sprintf("Administrative Interface Exposed: %s (Port %d)", service, port),
				Description: fmt.Sprintf("Administrative interface on port %d is publicly accessible", port),
				Severity:    "high",
				Category:    "network",
				Component:   "firewall",
				Location:    fmt.Sprintf("localhost:%d", port),
				Status:      "open",
				Remediation: fmt.Sprintf("Restrict access to administrative interface on port %d to trusted IP addresses only", port),
				Discovered:  time.Now(),
				Updated:     time.Now(),
			})
		}
	}

	return vulnerabilities, nil
}

// checkHTTPRedirect checks if HTTP traffic is properly redirected to HTTPS
func (sas *SecurityAuditService) checkHTTPRedirect() *Vulnerability {
	// Create HTTP client with timeout and redirect policy
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow one redirect to check if it goes to HTTPS
			if len(via) >= 1 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	// Try to connect to localhost on port 80
	resp, err := client.Get("http://localhost")
	if err != nil {
		// If we can't connect, that's a different issue (port not open)
		return nil
	}
	defer resp.Body.Close()

	// Check if the response is a redirect to HTTPS
	if resp.StatusCode == 301 || resp.StatusCode == 302 {
		location := resp.Header.Get("Location")
		if location != "" && (location[:8] == "https://" || location[:2] == "//") {
			// Good - redirecting to HTTPS
			return nil
		}
	}

	// HTTP traffic is not properly redirected to HTTPS
	return &Vulnerability{
		ID:          generateVulnID(),
		Title:       "HTTP Traffic Not Redirected to HTTPS",
		Description: "HTTP traffic on port 80 is not being redirected to HTTPS, exposing traffic to potential interception",
		Severity:    "medium",
		Category:    "network",
		Component:   "firewall",
		Location:    "localhost:80",
		Status:      "open",
		Remediation: "Configure web server to redirect all HTTP traffic to HTTPS (status code 301 or 302 with Location header pointing to https://)",
		Discovered:  time.Now(),
		Updated:     time.Now(),
	}
}