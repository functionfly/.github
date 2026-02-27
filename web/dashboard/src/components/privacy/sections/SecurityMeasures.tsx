export function SecurityMeasures() {
  return (
    <div>
      <h3 className="text-lg font-semibold mb-2">Security Measures</h3>
      <p className="text-muted-foreground mb-3">
        We implement comprehensive security measures to protect your personal information against unauthorized access, alteration, disclosure, or destruction:
      </p>
      <div className="space-y-3">
        <div>
          <h4 className="font-medium mb-1">Encryption</h4>
          <ul className="list-disc list-inside text-sm text-muted-foreground space-y-1">
            <li>Data in transit: TLS 1.3 encryption for all data transmissions</li>
            <li>Data at rest: AES-256 encryption for stored data</li>
            <li>Database encryption: Transparent data encryption for sensitive information</li>
          </ul>
        </div>
        <div>
          <h4 className="font-medium mb-1">Access Controls</h4>
          <ul className="list-disc list-inside text-sm text-muted-foreground space-y-1">
            <li>Role-based access control (RBAC) limiting data access to authorized personnel</li>
            <li>Multi-factor authentication (MFA) for all administrative access</li>
            <li>Regular access reviews and automated deprovisioning</li>
            <li>Principle of least privilege applied to all systems</li>
          </ul>
        </div>
        <div>
          <h4 className="font-medium mb-1">Security Monitoring</h4>
          <ul className="list-disc list-inside text-sm text-muted-foreground space-y-1">
            <li>24/7 security monitoring and intrusion detection systems</li>
            <li>Regular security audits and vulnerability assessments</li>
            <li>Automated threat detection and response systems</li>
            <li>Security information and event management (SIEM) systems</li>
          </ul>
        </div>
        <div>
          <h4 className="font-medium mb-1">Breach Response Procedures</h4>
          <p className="text-sm text-muted-foreground">
            In the event of a security breach, we have established incident response procedures that include immediate containment, investigation, notification to affected users within 72 hours (when legally required), and remediation measures. We maintain comprehensive breach logs and conduct post-incident reviews to prevent future occurrences.
          </p>
        </div>
      </div>
      <p className="text-sm text-muted-foreground mt-3">
        While we implement robust security measures, no system is completely immune to risks. We regularly update our security practices based on emerging threats and industry best practices.
      </p>
    </div>
  );
}