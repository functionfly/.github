import { studioProjectsApi } from '@/api/studioProjects';
import { useActiveEnvironment } from '@/hooks/useActiveEnvironment';
import { createDefaultProject, type StudioProjectsState } from '@/lib/studio-projects';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useCallback, useEffect, useMemo, useRef } from 'react';
import { toast } from 'sonner';

const WORKSPACE_QUERY_KEY = 'studio-workspace';

export interface UseStudioProjectsOptions {
  defaultStarterContent?: string;
}

function emptyState(starterContent: string): StudioProjectsState {
  const project = createDefaultProject('Untitled Project', starterContent);
  return {
    projects: [project],
    activeProjectId: project.id,
    activeFileId: project.files[0]?.id ?? null,
  };
}

export function useStudioProjects(options: UseStudioProjectsOptions = {}) {
  const { environment } = useActiveEnvironment();
  const queryClient = useQueryClient();
  const starterContent = options.defaultStarterContent ?? '';
  const saveContentTimerRef = useRef<number | null>(null);
  const pendingContentRef = useRef<{ fileId: string; content: string } | null>(null);

  const queryKey = useMemo(() => [WORKSPACE_QUERY_KEY, environment], [environment]);

  const { data, isLoading, error, refetch } = useQuery({
    queryKey,
    queryFn: () => studioProjectsApi.getWorkspace(),
    staleTime: 1000 * 30,
    retry: 1,
  });

  const state = data ?? (starterContent ? emptyState(starterContent) : null);

  const activeProject = useMemo(
    () => state?.projects.find((p) => p.id === state.activeProjectId) ?? null,
    [state]
  );

  const activeFile = useMemo(() => {
    if (!activeProject || !state?.activeFileId) return null;
    return activeProject.files.find((f) => f.id === state.activeFileId) ?? null;
  }, [activeProject, state?.activeFileId]);

  const patchWorkspaceCache = useCallback(
    (updater: (prev: StudioProjectsState) => StudioProjectsState) => {
      queryClient.setQueryData<StudioProjectsState>(queryKey, (prev) => {
        if (!prev) return prev;
        return updater(prev);
      });
    },
    [queryClient, queryKey]
  );

  const persistSession = useCallback(
    async (activeProjectId: string | null, activeFileId: string | null) => {
      try {
        await studioProjectsApi.saveSession(activeProjectId, activeFileId);
      } catch {
        // Non-fatal; workspace still works locally until next reload
      }
    },
    []
  );

  const createProjectMutation = useMutation({
    mutationFn: (name: string) =>
      studioProjectsApi.createProject(name, starterContent || undefined),
    onSuccess: (project) => {
      patchWorkspaceCache((prev) => ({
        projects: [project, ...prev.projects],
        activeProjectId: project.id,
        activeFileId: project.files[0]?.id ?? null,
      }));
      void persistSession(project.id, project.files[0]?.id ?? null);
      toast.success('Project created');
    },
    onError: (err: Error) => toast.error(err.message || 'Failed to create project'),
  });

  const deleteProjectMutation = useMutation({
    mutationFn: (projectId: string) => studioProjectsApi.deleteProject(projectId),
    onSuccess: async (_, projectId) => {
      const prev = queryClient.getQueryData<StudioProjectsState>(queryKey);
      if (!prev) {
        await refetch();
        return;
      }
      const remaining = prev.projects.filter((p) => p.id !== projectId);
      if (!remaining.length) {
        await refetch();
        toast.success('Project deleted');
        return;
      }
      const nextActive =
        prev.activeProjectId === projectId ? remaining[0].id : prev.activeProjectId;
      const nextProject = remaining.find((p) => p.id === nextActive) ?? remaining[0];
      const nextFileId =
        prev.activeProjectId === projectId ? (nextProject.files[0]?.id ?? null) : prev.activeFileId;
      patchWorkspaceCache(() => ({
        projects: remaining,
        activeProjectId: nextActive,
        activeFileId: nextFileId,
      }));
      await persistSession(nextActive, nextFileId);
      toast.success('Project deleted');
    },
    onError: (err: Error) => toast.error(err.message || 'Failed to delete project'),
  });

  const createFileMutation = useMutation({
    mutationFn: ({ projectId, name }: { projectId: string; name: string }) =>
      studioProjectsApi.createFile(projectId, name, { content: starterContent || undefined }),
    onSuccess: (file, { projectId }) => {
      patchWorkspaceCache((prev) => ({
        ...prev,
        projects: prev.projects.map((p) =>
          p.id === projectId ? { ...p, files: [...p.files, file] } : p
        ),
        activeFileId: file.id,
      }));
      void persistSession(projectId, file.id);
    },
    onError: (err: Error) => toast.error(err.message || 'Failed to create file'),
  });

  const deleteFileMutation = useMutation({
    mutationFn: ({ projectId, fileId }: { projectId: string; fileId: string }) =>
      studioProjectsApi.deleteFile(projectId, fileId),
    onSuccess: (_, { projectId, fileId }) => {
      patchWorkspaceCache((prev) => {
        const project = prev.projects.find((p) => p.id === projectId);
        const files = project?.files.filter((f) => f.id !== fileId) ?? [];
        const fallback = files[0]?.id ?? null;
        const nextFileId = prev.activeFileId === fileId ? fallback : prev.activeFileId;
        void persistSession(prev.activeProjectId, nextFileId);
        return {
          ...prev,
          projects: prev.projects.map((p) => (p.id === projectId ? { ...p, files } : p)),
          activeFileId: nextFileId,
        };
      });
    },
    onError: (err: Error) => toast.error(err.message || 'Failed to delete file'),
  });

  const saveFileMutation = useMutation({
    mutationFn: ({
      projectId,
      fileId,
      content,
    }: {
      projectId: string;
      fileId: string;
      content: string;
    }) => studioProjectsApi.updateFile(projectId, fileId, { content }),
  });

  const renameProjectMutation = useMutation({
    mutationFn: ({ projectId, name }: { projectId: string; name: string }) =>
      studioProjectsApi.updateProject(projectId, name),
    onSuccess: (project) => {
      patchWorkspaceCache((prev) => ({
        ...prev,
        projects: prev.projects.map((p) => (p.id === project.id ? project : p)),
      }));
    },
  });

  const duplicateProjectMutation = useMutation({
    mutationFn: (projectId: string) => studioProjectsApi.duplicateProject(projectId),
    onSuccess: (project) => {
      patchWorkspaceCache((prev) => ({
        projects: [project, ...prev.projects],
        activeProjectId: project.id,
        activeFileId: project.files[0]?.id ?? null,
      }));
      void persistSession(project.id, project.files[0]?.id ?? null);
      toast.success('Project duplicated');
    },
    onError: (err: Error) => toast.error(err.message || 'Failed to duplicate project'),
  });

  const createProject = useCallback(
    (name: string) => {
      createProjectMutation.mutate(name);
    },
    [createProjectMutation]
  );

  const openProject = useCallback(
    (projectId: string) => {
      patchWorkspaceCache((prev) => {
        const project = prev.projects.find((p) => p.id === projectId);
        if (!project) return prev;
        const fileId = project.files[0]?.id ?? null;
        void persistSession(projectId, fileId);
        return { ...prev, activeProjectId: projectId, activeFileId: fileId };
      });
    },
    [patchWorkspaceCache, persistSession]
  );

  const deleteProject = useCallback(
    (projectId: string) => deleteProjectMutation.mutate(projectId),
    [deleteProjectMutation]
  );

  const renameProject = useCallback(
    (projectId: string, name: string) => {
      renameProjectMutation.mutate({ projectId, name });
    },
    [renameProjectMutation]
  );

  const openFile = useCallback(
    (fileId: string) => {
      patchWorkspaceCache((prev) => {
        void persistSession(prev.activeProjectId, fileId);
        return { ...prev, activeFileId: fileId };
      });
    },
    [patchWorkspaceCache, persistSession]
  );

  const createFile = useCallback(
    (fileName: string) => {
      if (!activeProject) return null;
      createFileMutation.mutate({ projectId: activeProject.id, name: fileName });
      return null;
    },
    [activeProject, createFileMutation]
  );

  const deleteFile = useCallback(
    (fileId: string) => {
      if (!activeProject) return;
      deleteFileMutation.mutate({ projectId: activeProject.id, fileId });
    },
    [activeProject, deleteFileMutation]
  );

  const flushPendingContent = useCallback(() => {
    if (!activeProject || !pendingContentRef.current) return;
    const { fileId, content } = pendingContentRef.current;
    pendingContentRef.current = null;
    saveFileMutation.mutate({ projectId: activeProject.id, fileId, content });
  }, [activeProject, saveFileMutation]);

  const updateFileContent = useCallback(
    (fileId: string, content: string) => {
      patchWorkspaceCache((prev) => ({
        ...prev,
        projects: prev.projects.map((p) =>
          p.id === prev.activeProjectId
            ? {
                ...p,
                files: p.files.map((f) => (f.id === fileId ? { ...f, content } : f)),
              }
            : p
        ),
      }));

      pendingContentRef.current = { fileId, content };
      if (saveContentTimerRef.current) {
        window.clearTimeout(saveContentTimerRef.current);
      }
      saveContentTimerRef.current = window.setTimeout(() => {
        flushPendingContent();
      }, 800);
    },
    [patchWorkspaceCache, flushPendingContent]
  );

  const saveActiveFile = useCallback(() => {
    if (saveContentTimerRef.current) {
      window.clearTimeout(saveContentTimerRef.current);
      saveContentTimerRef.current = null;
    }
    flushPendingContent();
  }, [flushPendingContent]);

  useEffect(() => {
    return () => {
      if (saveContentTimerRef.current) {
        window.clearTimeout(saveContentTimerRef.current);
      }
      flushPendingContent();
    };
  }, [flushPendingContent]);

  const duplicateProject = useCallback(
    (projectId: string) => {
      duplicateProjectMutation.mutate(projectId);
    },
    [duplicateProjectMutation]
  );

  return {
    projects: state?.projects ?? [],
    activeProject,
    activeFile,
    isLoading,
    error,
    createProject,
    openProject,
    deleteProject,
    renameProject,
    duplicateProject,
    createFile,
    openFile,
    deleteFile,
    updateFileContent,
    saveActiveFile,
    isSaving: saveFileMutation.isPending,
  };
}
