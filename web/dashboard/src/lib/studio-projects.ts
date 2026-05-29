export interface StudioFile {
  id: string;
  name: string;
  path: string;
  content: string;
  language: string;
}

export interface StudioProject {
  id: string;
  name: string;
  files: StudioFile[];
  createdAt: string;
  updatedAt: string;
}

export interface StudioProjectsState {
  projects: StudioProject[];
  activeProjectId: string | null;
  activeFileId: string | null;
}

export function createFileId(): string {
  return `file-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
}

export function createProjectId(): string {
  return `project-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
}

export function inferLanguage(fileName: string): string {
  const ext = fileName.split('.').pop()?.toLowerCase();
  switch (ext) {
    case 'ts':
    case 'tsx':
      return 'typescript';
    case 'js':
    case 'jsx':
      return 'javascript';
    case 'py':
      return 'python';
    case 'go':
      return 'go';
    case 'rs':
      return 'rust';
    case 'json':
      return 'json';
    case 'md':
      return 'markdown';
    default:
      return 'plaintext';
  }
}

export function defaultMainFile(content: string): StudioFile {
  return {
    id: createFileId(),
    name: 'main.ts',
    path: 'src/main.ts',
    content,
    language: 'typescript',
  };
}

export function createDefaultProject(name: string, starterContent: string): StudioProject {
  const now = new Date().toISOString();
  const mainFile = defaultMainFile(starterContent);
  return {
    id: createProjectId(),
    name,
    files: [mainFile],
    createdAt: now,
    updatedAt: now,
  };
}

export function storageKey(userId: string, environment: string): string {
  return `functionfly:studio:projects:${userId}:${environment}`;
}

export function loadProjectsState(key: string): StudioProjectsState | null {
  if (typeof window === 'undefined') return null;
  try {
    const raw = window.localStorage.getItem(key);
    if (!raw) return null;
    return JSON.parse(raw) as StudioProjectsState;
  } catch {
    return null;
  }
}

export function saveProjectsState(key: string, state: StudioProjectsState): void {
  if (typeof window === 'undefined') return;
  window.localStorage.setItem(key, JSON.stringify(state));
}

export function groupFilesByDirectory(files: StudioFile[]): Map<string, StudioFile[]> {
  const groups = new Map<string, StudioFile[]>();
  for (const file of files) {
    const parts = file.path.split('/');
    const dir = parts.length > 1 ? parts.slice(0, -1).join('/') : '';
    const existing = groups.get(dir) ?? [];
    existing.push(file);
    groups.set(dir, existing);
  }
  return new Map([...groups.entries()].sort(([a], [b]) => a.localeCompare(b)));
}
