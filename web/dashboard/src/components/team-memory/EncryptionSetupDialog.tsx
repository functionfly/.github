import { useState } from 'react';
import { Shield, Key, Eye, EyeOff, Copy, Check } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import {
  generateSecurePassphrase,
  storePassphrase,
  hasPassphrase,
  getPassphrase,
  clearPassphrase,
  initTeamMemoryCrypto,
} from '@/utils/team-memory-crypto';

interface EncryptionSetupDialogProps {
  teamId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function EncryptionSetupDialog({
  teamId,
  open,
  onOpenChange,
}: EncryptionSetupDialogProps) {
  const [activeTab, setActiveTab] = useState('status');
  const [generatedPassphrase, setGeneratedPassphrase] = useState('');
  const [storedPassphrase, setStoredPassphrase] = useState('');
  const [inputPassphrase, setInputPassphrase] = useState('');
  const [showPassphrase, setShowPassphrase] = useState(false);
  const [copied, setCopied] = useState(false);
  const [isVerified, setIsVerified] = useState(false);
  const [error, setError] = useState('');

  const hasStoredPassphrase = hasPassphrase(teamId);

  const handleGenerate = () => {
    const passphrase = generateSecurePassphrase(16);
    setGeneratedPassphrase(passphrase);
    setStoredPassphrase(passphrase);
  };

  const handleCopy = () => {
    navigator.clipboard.writeText(storedPassphrase);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleStore = () => {
    if (storedPassphrase) {
      storePassphrase(teamId, storedPassphrase);
      setActiveTab('status');
    }
  };

  const handleUnlock = async () => {
    setError('');
    try {
      await initTeamMemoryCrypto(teamId, inputPassphrase);
      setIsVerified(true);
      setTimeout(() => {
        setIsVerified(false);
        setInputPassphrase('');
        setActiveTab('status');
      }, 2000);
    } catch (err) {
      setError('Invalid passphrase. Please check and try again.');
    }
  };

  const handleClear = () => {
    clearPassphrase(teamId);
    setActiveTab('status');
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Shield className="h-5 w-5" />
            Encryption Setup
          </DialogTitle>
          <DialogDescription>
            Manage client-side encryption for team memories. Your passphrase is
            never sent to the server.
          </DialogDescription>
        </DialogHeader>

        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <TabsList className="grid w-full grid-cols-3">
            <TabsTrigger value="status">Status</TabsTrigger>
            <TabsTrigger value="setup">Set Passphrase</TabsTrigger>
            <TabsTrigger value="unlock">Unlock</TabsTrigger>
          </TabsList>

          <TabsContent value="status" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Encryption Status</CardTitle>
                <CardDescription>
                  Current state of encryption for this session
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex items-center justify-between">
                  <span className="text-sm">Passphrase stored:</span>
                  <Badge variant={hasStoredPassphrase ? 'default' : 'secondary'}>
                    {hasStoredPassphrase ? 'Yes' : 'No'}
                  </Badge>
                </div>

                <div className="flex items-center justify-between">
                  <span className="text-sm">Session key ready:</span>
                  <Badge variant={isVerified ? 'default' : 'secondary'}>
                    {isVerified ? 'Ready' : 'Not initialized'}
                  </Badge>
                </div>

                <Alert>
                  <Shield className="h-4 w-4" />
                  <AlertDescription>
                    Passphrases are stored only in sessionStorage and are lost when
                    you close the browser tab. Team members must share the passphrase
                    out-of-band to access encrypted memories.
                  </AlertDescription>
                </Alert>

                {hasStoredPassphrase && (
                  <Button
                    variant="destructive"
                    onClick={handleClear}
                    className="w-full"
                  >
                    Clear Passphrase from Session
                  </Button>
                )}
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="setup" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Set Team Passphrase</CardTitle>
                <CardDescription>
                  Generate or enter a passphrase for encrypting team memories
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label>Passphrase</Label>
                  <div className="flex gap-2">
                    <div className="relative flex-1">
                      <Input
                        type={showPassphrase ? 'text' : 'password'}
                        value={storedPassphrase}
                        onChange={(e) => setStoredPassphrase(e.target.value)}
                        placeholder="Enter or generate a passphrase"
                      />
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        className="absolute right-2 top-1/2 -translate-y-1/2"
                        onClick={() => setShowPassphrase(!showPassphrase)}
                      >
                        {showPassphrase ? (
                          <EyeOff className="h-4 w-4" />
                        ) : (
                          <Eye className="h-4 w-4" />
                        )}
                      </Button>
                    </div>
                    {storedPassphrase && (
                      <Button
                        variant="outline"
                        size="icon"
                        onClick={handleCopy}
                      >
                        {copied ? (
                          <Check className="h-4 w-4" />
                        ) : (
                          <Copy className="h-4 w-4" />
                        )}
                      </Button>
                    )}
                  </div>
                </div>

                <Button
                  variant="secondary"
                  onClick={handleGenerate}
                  className="w-full"
                >
                  <Key className="h-4 w-4 mr-2" />
                  Generate Secure Passphrase
                </Button>

                {generatedPassphrase && (
                  <Alert className="bg-amber-50 border-amber-200">
                    <AlertDescription className="text-amber-800">
                      <strong>Important:</strong> Save this passphrase! It cannot be
                      recovered. All team members need this to access encrypted
                      memories.
                    </AlertDescription>
                  </Alert>
                )}

                <Button
                  onClick={handleStore}
                  disabled={!storedPassphrase}
                  className="w-full"
                >
                  Store Passphrase in Session
                </Button>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="unlock" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Unlock Encryption</CardTitle>
                <CardDescription>
                  Enter your team passphrase to decrypt memories
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                {error && (
                  <Alert variant="destructive">
                    <AlertDescription>{error}</AlertDescription>
                  </Alert>
                )}

                <div className="space-y-2">
                  <Label>Team Passphrase</Label>
                  <div className="relative">
                    <Input
                      type={showPassphrase ? 'text' : 'password'}
                      value={inputPassphrase}
                      onChange={(e) => setInputPassphrase(e.target.value)}
                      placeholder="Enter team passphrase"
                    />
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      className="absolute right-2 top-1/2 -translate-y-1/2"
                      onClick={() => setShowPassphrase(!showPassphrase)}
                    >
                      {showPassphrase ? (
                        <EyeOff className="h-4 w-4" />
                      ) : (
                        <Eye className="h-4 w-4" />
                      )}
                    </Button>
                  </div>
                </div>

                <Button
                  onClick={handleUnlock}
                  disabled={!inputPassphrase}
                  className="w-full"
                >
                  <Shield className="h-4 w-4 mr-2" />
                  Verify & Unlock
                </Button>

                {isVerified && (
                  <Alert className="bg-green-50 border-green-200">
                    <AlertDescription className="text-green-800">
                      Passphrase verified! Encryption is ready.
                    </AlertDescription>
                  </Alert>
                )}

                <Alert>
                  <AlertDescription>
                    The passphrase will be stored in sessionStorage for this
                    browser session only. You&apos;ll need to re-enter it if you
                    refresh or close the tab.
                  </AlertDescription>
                </Alert>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  );
}

import { Badge } from '@/components/ui/badge';
