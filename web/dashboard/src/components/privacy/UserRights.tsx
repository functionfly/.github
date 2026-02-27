export function UserRights() {
  return (
    <div className="privacy-card">
      <div className="space-y-6">
        <div>
          <h3 className="text-xl font-bold text-white">Your Rights</h3>
        </div>
        <div className="space-y-4">
          <div>
            <h4 className="font-semibold mb-2">GDPR Rights (EU Users)</h4>
            <ul className="list-disc list-inside text-muted-foreground space-y-1">
              <li>Right to access your personal data</li>
              <li>Right to rectification of inaccurate data</li>
              <li>Right to erasure ("right to be forgotten")</li>
              <li>Right to restrict processing</li>
              <li>Right to data portability</li>
              <li>Right to object to processing</li>
            </ul>
          </div>

          <div>
            <h4 className="font-semibold mb-2">CCPA Rights (California Users)</h4>
            <ul className="list-disc list-inside text-muted-foreground space-y-1">
              <li>Right to know what personal information is collected</li>
              <li>Right to know if personal information is sold or shared</li>
              <li>Right to opt-out of the sale of personal information</li>
              <li>Right to delete personal information</li>
              <li>Right to non-discrimination for exercising CCPA rights</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
}