import React from 'react';
import { Button, Input } from '@functionfly/ui-core';
import { ChevronDown, ChevronRight, FileCode, FolderOpen, Plus, Trash2 } from 'lucide-react';
import { useMemo, useState } from 'react';
import { groupFilesByDirectory, type StudioFile, type StudioProject } from '@/lib/studio-projects';

interface FilesPanelProps {
  project: StudioProject | null;
  activeFileId: string | null;
  onOpenFile: (fileId: string) => void;
  onCreateFile: (name: string) => void;
  onDeleteFile: (fileId: string) => void;
  onOpenProject: () => void;
  onNewProject: () => void;
}

export function FilesPanel({
  project,
  activeFileId,
  onOpenFile,
  onCreateFile,
  onDeleteFile,
  onOpenProject,
  onNewProject,
}: FilesPanelProps) {
  const [expandedDirs, setExpandedDirs] = useState<Set<string>>(() => new Set(['src', '']));
  const [newFileName, setNewFileName] = useState('');
  const [showNewFile, setShowNewFile] = useState(false);

  const grouped = useMemo(
    () => groupFilesByDirectory(project?.files ?? []),
    [project?.files]
  );

  const toggleDir = (dir: string) => {
    setExpandedDirs((prev) => {
      const next = new Set(prev);
      if (next.has(dir)) next.delete(dir);
      else next.add(dir);
      return next;
    });
  };

  const handleCreateFile = () => {
    const name = newFileName.trim();
    if (!name) return;
    onCreateFile(name);
    setNewFileName('');
    setShowNewFile(false);
  };

  if (!project) {
    return (
      <div className="p-4 text-center space-y-3">
        <p className="text-sm text-text-muted">No project open</p>
        <Button size="sm" onClick={onNewProject}>
          New Project
        </Button>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      <div className="px-3 py-2 border-b border-border-subtle space-y-2">
        <div className="flex items-center gap-2 min-w-0">
          <FolderOpen className="size-4 shrink-0 text-brand-400" />
          <span className="text-sm font-medium truncate">{project.name}</span>
        </div>
        <div className="flex gap-1">
          <Button variant="ghost" size="sm" className="flex-1 h-7 text-[11px]" onClick={onOpenProject}>
            Open
          </Button>
          <Button variant="ghost" size="sm" className="flex-1 h-7 text-[11px]" onClick={onNewProject}>
            New
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className="h-7 px-2"
            title="New file"
            onClick={() => setShowNewFile((v) => !v)}
          >
            <Plus className="size-3.5" />
          </Button>
        </div>
        {showNewFile && (
          <div className="flex gap-1">
            <Input
              value={newFileName}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setNewFileName(e.target.value)}
              placeholder="handler.ts"
              className="h-7 text-xs"
              autoFocus
              onKeyDown={(e: React.KeyboardEvent<HTMLInputElement>) => {
                if (e.key === 'Enter') handleCreateFile();
                if (e.key === 'Escape') setShowNewFile(false);
              }}
            />
            <Button size="sm" className="h-7 px-2 text-xs" onClick={handleCreateFile}>
              Add
            </Button>
          </div>
        )}
      </div>

      <div className="flex-1 overflow-y-auto py-1">
        {[...grouped.entries()].map(([dir, files]) => {
          const label = dir || '(root)';
          const isExpanded = expandedDirs.has(dir);
          return (
            <div key={dir || '__root'}>
              <button
                className="w-full flex items-center gap-1 px-2 py-1 text-[11px] text-text-muted hover:text-text-primary hover:bg-bg-hover"
                onClick={() => toggleDir(dir)}
              >
                {isExpanded ? (
                  <ChevronDown className="size-3 shrink-0" />
                ) : (
                  <ChevronRight className="size-3 shrink-0" />
                )}
                <span className="truncate">{label}</span>
              </button>
              {isExpanded &&
                files.map((file) => (
                  <FileRow
                    key={file.id}
                    file={file}
                    isActive={file.id === activeFileId}
                    onOpen={() => onOpenFile(file.id)}
                    onDelete={() => onDeleteFile(file.id)}
                  />
                ))}
            </div>
          );
        })}
      </div>
    </div>
  );
}

function FileRow({
  file,
  isActive,
  onOpen,
  onDelete,
}: {
  file: StudioFile;
  isActive: boolean;
  onOpen: () => void;
  onDelete: () => void;
}) {
  return (
    <div
      className={`group flex items-center gap-1 pl-5 pr-2 py-1 cursor-pointer ${
        isActive ? 'bg-brand-500/15 text-brand-400' : 'hover:bg-bg-hover text-text-secondary'
      }`}
    >
      <button className="flex-1 flex items-center gap-1.5 min-w-0 text-left" onClick={onOpen}>
        <FileCode className="size-3.5 shrink-0" />
        <span className="text-xs truncate">{file.name}</span>
      </button>
      <button
        className="p-0.5 rounded opacity-0 group-hover:opacity-100 text-text-muted hover:text-error"
        title="Delete file"
        onClick={(e) => {
          e.stopPropagation();
          onDelete();
        }}
      >
        <Trash2 className="size-3" />
      </button>
    </div>
  );
}
