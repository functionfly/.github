import { useState } from 'react';
import { CodePasteModal } from '../components/CodePaste';
import { useNavigate } from 'react-router-dom';
import { PageGrid } from '@/components/containment';
import './PasteCodePage.css';

export function PasteCodePage() {
  const [isModalOpen, setIsModalOpen] = useState(true);
  const navigate = useNavigate();

  const handleClose = () => {
    setIsModalOpen(false);
    navigate(-1);
  };

  const handleImportComplete = (functions: Array<{ id: string; name: string }>) => {
    console.log('Imported functions:', functions);
  };

  return (
    <div className="paste-code-page">
      <PageGrid />
      <CodePasteModal
        isOpen={isModalOpen}
        onClose={handleClose}
        onImportComplete={handleImportComplete}
      />
    </div>
  );
}
