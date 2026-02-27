export function DataRetention() {
  return (
    <div>
      <h3 className="text-lg font-semibold mb-2">Data Retention</h3>
      <p className="text-muted-foreground mb-3">
        We retain your personal information for different periods of time depending on the type of data and the purpose for which it was collected:
      </p>
      <div className="space-y-3">
        <div>
          <h4 className="font-medium mb-1">Account Data</h4>
          <p className="text-sm text-muted-foreground">
            Account information, profile data, and authentication details are retained for as long as your account remains active and for up to 3 years after account deletion to comply with legal obligations and resolve disputes.
          </p>
        </div>
        <div>
          <h4 className="font-medium mb-1">Usage and Analytics Data</h4>
          <p className="text-sm text-muted-foreground">
            Usage analytics, performance metrics, and log data are typically retained for 2 years to help us improve our services and maintain system security.
          </p>
        </div>
        <div>
          <h4 className="font-medium mb-1">Communication Records</h4>
          <p className="text-sm text-muted-foreground">
            Customer support communications and service-related messages are retained for 3 years for quality assurance and legal compliance purposes.
          </p>
        </div>
        <div>
          <h4 className="font-medium mb-1">Marketing Data</h4>
          <p className="text-sm text-muted-foreground">
            Marketing preferences and campaign interaction data are retained for 1 year after your last interaction with our marketing communications.
          </p>
        </div>
      </div>
      <p className="text-sm text-muted-foreground mt-3">
        We may retain data longer when required by law, involved in legal proceedings, or necessary for legitimate business purposes. You can request deletion of your data at any time.
      </p>
    </div>
  );
}