import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { ChevronDown, ChevronUp, AlertCircle, HelpCircle, BookOpen } from 'lucide-react';
import { securityApi } from '@/api/security';
import type { SecurityFAQ as SecurityFAQType } from '../types';

interface SecurityFAQProps {
  expandedSection: string | null;
  toggleSection: (sectionId: string) => void;
}

export function SecurityFAQ({ expandedSection, toggleSection }: SecurityFAQProps) {
  const [faqs, setFaqs] = useState<SecurityFAQType[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchFAQs = async () => {
      try {
        setLoading(true);
        const response = await securityApi.getSecurityFAQ();
        setFaqs(response.faqs);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load security FAQs');
      } finally {
        setLoading(false);
      }
    };

    fetchFAQs();
  }, []);

  if (loading) {
    return (
      <div className="space-y-4">
          {[1, 2, 3, 4, 5].map((i) => (
            <div key={i} className="border rounded-lg">
              <div className="flex items-center justify-between p-4">
                <div className="h-4 bg-muted animate-pulse rounded w-3/4"></div>
                <div className="h-4 w-4 bg-muted animate-pulse rounded"></div>
              </div>
            </div>
          ))}
      </div>
    );
  }

  if (error) {
    return (
          <div className="flex items-center justify-center py-8 text-center">
            <div>
              <AlertCircle className="h-8 w-8 text-red-500 mx-auto mb-2" />
              <p className="text-sm text-muted-foreground">{error}</p>
            </div>
          </div>
    );
  }

  return (
    <div className="space-y-4">
        {faqs.map((faq) => (
          <div key={faq.id} className="border rounded-lg overflow-hidden">
            <button
              onClick={() => toggleSection(faq.id)}
              className="w-full flex items-center justify-between p-4 text-left hover:bg-muted/50 active:bg-muted/70 transition-colors touch-manipulation md:p-4 p-6"
              style={{ minHeight: '60px' }} // Ensure minimum touch target size
            >
              <span className="font-medium text-sm md:text-base pr-3 flex-1">{faq.question}</span>
              <div className="shrink-0 ml-2">
                {expandedSection === faq.id ? (
                  <ChevronUp className="h-5 w-5 md:h-4 md:w-4 transition-transform duration-200" />
                ) : (
                  <ChevronDown className="h-5 w-5 md:h-4 md:w-4 transition-transform duration-200" />
                )}
              </div>
            </button>
            <div
              className={`transition-all duration-300 ease-in-out ${
                expandedSection === faq.id
                  ? 'max-h-96 opacity-100'
                  : 'max-h-0 opacity-0 overflow-hidden'
              }`}
            >
              <div className="px-4 pb-4 md:px-4 md:pb-4 px-6 pb-6">
                <p className="text-muted-foreground text-sm md:text-base leading-relaxed">{faq.answer}</p>
              </div>
            </div>
          </div>
        ))}
    </div>
  );
}