import type { CollabEvent } from '@/api/studioCollab';
import { useStudioCollabEvents } from '@/hooks/useStudioCollab';
import { useMemo } from 'react';

function filterByType(events: CollabEvent[] | undefined, type: string) {
  return (events ?? []).filter((ev) => ev.event_type === type);
}

export function useStudioCollabData() {
  const { data: allEventsData, isLoading } = useStudioCollabEvents({ limit: 200 });

  const events = allEventsData?.events ?? [];

  const promptVersionsData = useMemo(
    () =>
      filterByType(events, 'prompt_version').map((ev) => ({
        id: ev.id,
        metadata: ev.metadata as {
          prompt?: string;
          user_name?: string;
          user_color?: string;
          changes?: string;
        },
        created_at: ev.created_at,
      })),
    [events]
  );

  const pairSessionsData = useMemo(
    () =>
      filterByType(events, 'pair_session').map((ev) => ({
        id: ev.id,
        metadata: ev.metadata as {
          host_name?: string;
          host_color?: string;
          guest_name?: string;
          guest_color?: string;
          status?: string;
          current_file?: string;
          current_line?: number;
        },
        created_at: ev.created_at,
      })),
    [events]
  );

  const commentsData = useMemo(
    () =>
      filterByType(events, 'comment').map((ev) => ({
        id: ev.id,
        metadata: ev.metadata as {
          user_name?: string;
          user_color?: string;
          content?: string;
          line?: number;
          resolved?: boolean;
        },
        created_at: ev.created_at,
      })),
    [events]
  );

  const annotationsData = useMemo(
    () =>
      filterByType(events, 'annotation').map((ev) => ({
        id: ev.id,
        metadata: ev.metadata as {
          user_name?: string;
          user_color?: string;
          target_id?: string;
          target_type?: string;
          content?: string;
          position?: { x: number; y: number };
          resolved?: boolean;
        },
        created_at: ev.created_at,
      })),
    [events]
  );

  const graphEditsData = useMemo(
    () =>
      filterByType(events, 'graph_edit').map((ev) => ({
        id: ev.id,
        created_by: ev.created_by,
        metadata: ev.metadata as {
          user_name?: string;
          node_id?: string;
          field?: string;
          old_value?: string;
          new_value?: string;
        },
        created_at: ev.created_at,
      })),
    [events]
  );

  const conflictsData = useMemo(
    () =>
      filterByType(events, 'conflict')
        .filter((ev) => !ev.metadata?.resolution)
        .map((ev) => ({
          id: ev.id,
          metadata: ev.metadata as {
            field?: string;
            current_user?: string;
            current_value?: string;
            incoming_user?: string;
            incoming_value?: string;
          },
        })),
    [events]
  );

  return {
    isLoading,
    promptVersionsData,
    pairSessionsData,
    commentsData,
    annotationsData,
    graphEditsData,
    conflictsData,
  };
}
