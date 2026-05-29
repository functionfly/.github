import type { StudioProject } from '@/lib/studio-projects';
import { Button, Input } from '@functionfly/ui-core';
import { FolderOpen, Plus, Trash2 } from 'lucide-react';
import React, { useState } from 'react';

interface ProjectDialogProps {
  open: boolean;
  mode: 'open' | 'create';
  projects: StudioProject[];
  activeProjectId: string | null;
  onClose: () => void;
  onOpen: (projectId: string) => void;
  onCreate: (name: string) => void;
  onDelete: (projectId: string) => void;
}

export function ProjectDialog({
  open,
  mode,
  projects,
  activeProjectId,
  onClose,
  onOpen,
  onCreate,
  onDelete,
}: ProjectDialogProps) {
  const [newName, setNewName] = useState('');

  if (!open) return null;

  const handleCreate = () => {
    onCreate(newName.trim() || 'Untitled Project');
    setNewName('');
    onClose();
  };

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/50 backdrop-blur-sm">
      <div className="w-full max-w-md rounded-xl border border-border-subtle bg-bg-primary shadow-2xl">
        <div className="flex items-center justify-between px-4 py-3 border-b border-border-subtle">
          <h2 className="text-sm font-semibold text-text-primary">
            {mode === 'create' ? 'New Project' : 'Open Project'}
          </h2>
          <button onClick={onClose} className="text-xs text-text-muted hover:text-text-primary">
            Esc
          </button>
        </div>

        {mode === 'create' ? (
          <div className="p-4 space-y-3">
            <label className="text-xs text-text-muted">Project name</label>
            <Input
              value={newName}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setNewName(e.target.value)}
              placeholder="My workflow project"
              autoFocus
              onKeyDown={(e: React.KeyboardEvent<HTMLInputElement>) => {
                if (e.key === 'Enter') handleCreate();
                if (e.key === 'Escape') onClose();
              }}
            />
            <div className="flex justify-end gap-2 pt-2">
              <Button variant="ghost" size="sm" onClick={onClose}>
                Cancel
              </Button>
              <Button size="sm" onClick={handleCreate}>
                Create
              </Button>
            </div>
          </div>
        ) : (
          <div className="p-2 max-h-80 overflow-y-auto">
            {projects.length === 0 ? (
              <p className="px-3 py-6 text-sm text-text-muted text-center">
                No projects yet. Use File → New Project to get started.
              </p>
            ) : (
              projects.map((project) => (
                <div
                  key={project.id}
                  className="flex items-center gap-2 px-2 py-2 rounded-lg hover:bg-bg-hover group"
                >
                  <button
                    className="flex-1 flex items-center gap-2 text-left min-w-0"
                    onClick={() => {
                      onOpen(project.id);
                      onClose();
                    }}
                  >
                    <FolderOpen className="size-4 shrink-0 text-brand-400" />
                    <div className="min-w-0">
                      <div className="text-sm text-text-primary truncate">{project.name}</div>
                      <div className="text-[10px] text-text-muted">
                        {project.files.length} file{project.files.length === 1 ? '' : 's'}
                        {activeProjectId === project.id ? ' · current' : ''}
                      </div>
                    </div>
                  </button>
                  <button
                    className="p-1.5 rounded opacity-0 group-hover:opacity-100 text-text-muted hover:text-error hover:bg-error/10 transition-all"
                    title="Delete project"
                    onClick={() => onDelete(project.id)}
                  >
                    <Trash2 className="size-3.5" />
                  </button>
                </div>
              ))
            )}
            <div className="p-2 border-t border-border-subtle mt-1">
              <Button
                variant="ghost"
                size="sm"
                className="w-full justify-start gap-2"
                onClick={() => {
                  onClose();
                  onCreate('Untitled Project');
                }}
              >
                <Plus className="size-4" />
                New project
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
