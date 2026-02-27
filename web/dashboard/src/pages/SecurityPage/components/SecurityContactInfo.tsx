import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { AlertCircle } from 'lucide-react';
import { securityApi } from '@/api/security';
import type { ContactInfo } from '../types';
import * as LucideIcons from 'lucide-react';

export function SecurityContactInfo() {
  const [contacts, setContacts] = useState<ContactInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchContacts = async () => {
      try {
        setLoading(true);
        const response = await securityApi.getContactInfo();
        setContacts(response.contacts);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load contact information');
      } finally {
        setLoading(false);
      }
    };

    fetchContacts();
  }, []);

  const getIcon = (iconName: string) => {
    const IconComponent = (LucideIcons as any)[iconName];
    return IconComponent || LucideIcons.Shield;
  };

  if (loading) {
    return (
      <div className="space-y-4">
        <div className="h-4 bg-muted animate-pulse rounded w-full mb-4"></div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {[1, 2].map((i) => (
              <div key={i} className="flex items-center gap-3">
                <div className="w-10 h-10 bg-muted animate-pulse rounded-full"></div>
                <div className="flex-1">
                  <div className="h-4 bg-muted animate-pulse rounded w-24 mb-1"></div>
                  <div className="h-3 bg-muted animate-pulse rounded w-32 mb-1"></div>
                  <div className="h-3 bg-muted animate-pulse rounded w-20"></div>
                </div>
              </div>
            ))}
          </div>
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
        <p className="text-muted-foreground">
          For security-related questions, vulnerability reports, or compliance inquiries,
          please contact our security team:
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {contacts.map((contact) => {
            const IconComponent = getIcon(contact.icon);
            return (
              <div key={contact.type} className="flex items-center gap-3">
                <div className={`w-10 h-10 bg-${contact.type === 'security' ? 'red' : 'blue'}-500/10 rounded-full flex items-center justify-center`}>
                  <IconComponent className={`h-5 w-5 text-${contact.type === 'security' ? 'red' : 'blue'}-500`} />
                </div>
                <div>
                  <h4 className="font-medium">{contact.title}</h4>
                  <p className="text-sm text-muted-foreground">{contact.email}</p>
                  <p className="text-xs text-muted-foreground">{contact.notes}</p>
                </div>
              </div>
            );
          })}
      </div>
    </div>
  );
}