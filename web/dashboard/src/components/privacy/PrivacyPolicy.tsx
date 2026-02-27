import { InformationWeCollect } from './sections/InformationWeCollect';
import { HowWeUseInformation } from './sections/HowWeUseInformation';
import { InformationSharing } from './sections/InformationSharing';
import { DataRetention } from './sections/DataRetention';
import { ThirdPartyServices } from './sections/ThirdPartyServices';
import { InternationalDataTransfers } from './sections/InternationalDataTransfers';
import { SecurityMeasures } from './sections/SecurityMeasures';
import { ChildrensPrivacy } from './sections/ChildrensPrivacy';

export function PrivacyPolicy() {
  return (
    <div className="privacy-card">
      <div className="space-y-6">
        <div>
          <h3 className="text-xl font-bold text-white">Privacy Policy</h3>
        </div>

        <InformationWeCollect />
        <HowWeUseInformation />
        <InformationSharing />
        <DataRetention />
        <ThirdPartyServices />
        <InternationalDataTransfers />
        <SecurityMeasures />
        <ChildrensPrivacy />
      </div>
    </div>
  );
}