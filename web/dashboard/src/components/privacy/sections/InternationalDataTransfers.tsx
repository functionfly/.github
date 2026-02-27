export function InternationalDataTransfers() {
  return (
    <div>
      <h3 className="text-lg font-semibold mb-2">International Data Transfers</h3>
      <p className="text-muted-foreground mb-3">
        FunctionFly operates globally, and your personal information may be transferred to and processed in countries other than your own. We implement appropriate safeguards to protect your data during international transfers:
      </p>
      <div className="space-y-3">
        <div>
          <h4 className="font-medium mb-1">Data Transfer Mechanisms</h4>
          <ul className="list-disc list-inside text-sm text-muted-foreground space-y-1">
            <li>Standard Contractual Clauses (SCCs) approved by the European Commission</li>
            <li>Adequacy decisions for transfers to countries with equivalent privacy protections</li>
            <li>Binding Corporate Rules for intra-group transfers</li>
            <li>Your explicit consent when required by applicable law</li>
          </ul>
        </div>
        <div>
          <h4 className="font-medium mb-1">Data Hosting Locations</h4>
          <p className="text-sm text-muted-foreground">
            Your data may be stored and processed in data centers located in the United States, European Union, and other jurisdictions. We use industry-leading cloud providers with robust security certifications.
          </p>
        </div>
        <div>
          <h4 className="font-medium mb-1">Cross-Border Data Flows</h4>
          <p className="text-sm text-muted-foreground">
            When transferring data across borders, we ensure that appropriate safeguards are in place, including encryption in transit and at rest, access controls, and regular security assessments.
          </p>
        </div>
      </div>
      <p className="text-sm text-muted-foreground mt-3">
        If you are located in the European Economic Area or other regions with strict data transfer requirements, you have the right to obtain details about the safeguards we use for international data transfers.
      </p>
    </div>
  );
}