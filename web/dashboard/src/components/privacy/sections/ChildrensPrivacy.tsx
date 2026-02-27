export function ChildrensPrivacy() {
  return (
    <div>
      <h3 className="text-lg font-semibold mb-2">Children's Privacy</h3>
      <p className="text-muted-foreground mb-3">
        FunctionFly is committed to protecting children's privacy online. Our services are not intended for children under 13 years of age.
      </p>
      <div className="space-y-3">
        <div>
          <h4 className="font-medium mb-1">COPPA Compliance</h4>
          <p className="text-sm text-muted-foreground">
            We comply with the Children's Online Privacy Protection Act (COPPA) and similar international laws regarding children's privacy. We do not knowingly collect personal information from children under 13 without verifiable parental consent.
          </p>
        </div>
        <div>
          <h4 className="font-medium mb-1">Age Verification</h4>
          <p className="text-sm text-muted-foreground">
            During account registration, users must confirm they are at least 13 years old. If we become aware that we have collected personal information from a child under 13, we will take steps to delete such information promptly.
          </p>
        </div>
        <div>
          <h4 className="font-medium mb-1">Parental Rights</h4>
          <p className="text-sm text-muted-foreground">
            Parents or guardians may request to review, modify, or delete personal information collected from their child under 13. To exercise these rights, please contact us using the information provided below.
          </p>
        </div>
        <div>
          <h4 className="font-medium mb-1">Content and Features</h4>
          <p className="text-sm text-muted-foreground">
            Our platform is designed for business and professional use. We do not offer features specifically targeted at children, and we do not collect information for marketing to children.
          </p>
        </div>
      </div>
      <p className="text-sm text-muted-foreground mt-3">
        If you believe we have collected information from a child under 13 without proper consent, please contact us immediately so we can address the issue.
      </p>
    </div>
  );
}