/**
 * MCP Center - MCP Settings Dialog
 * Modal wrapper for MCPSettingsPanel to enable inline editing
 */

import { useState } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog';
import { MCPSettingsPanel } from '@/components/registry/MCPSettingsPanel';
import type { MCPSettings } from '@/components/registry/MCPSettingsPanel';

interface MCPSettingsDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  author: string;
  name: string;
  onSaved?: (settings: MCPSettings) => void;
}

export function MCPSettingsDialog({
  open,
  onOpenChange,
  author,
  name,
  onSaved,
}: MCPSettingsDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>MCP Settings</DialogTitle>
          <DialogDescription>
            Configure Model Context Protocol settings for {author}/{name}
          </DialogDescription>
        </DialogHeader>
        <MCPSettingsPanel
          author={author}
          name={name}
          onSaved={onSaved}
        />
      </DialogContent>
    </Dialog>
  );
}