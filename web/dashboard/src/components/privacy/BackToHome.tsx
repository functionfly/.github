import { Link } from 'react-router-dom';
import { FileText } from 'lucide-react';

export function BackToHome() {
  return (
    <div className="text-center">
      <Link to="/">
        <button className="privacy-back-button">
          <FileText className="h-4 w-4 mr-2" />
          Back to Home
        </button>
      </Link>
    </div>
  );
}