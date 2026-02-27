export function InformationWeCollect() {
  return (
    <div>
      <h3 className="text-lg font-semibold mb-2">Information We Collect</h3>
      <p className="text-muted-foreground mb-3">
        We collect information you provide directly to us, such as when you create an account,
        use our services, or contact us for support.
      </p>
      <ul className="list-disc list-inside text-muted-foreground space-y-1">
        <li>Account information (email, name, company)</li>
        <li>Usage data and analytics</li>
        <li>Device and browser information</li>
        <li>IP address and location data</li>
      </ul>
    </div>
  );
}