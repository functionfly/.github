-- Bulk-delete old function_logs for storage cleanup. Uses SECURITY DEFINER so RLS
-- on function_logs does not block the orchestrator's application role.
CREATE OR REPLACE FUNCTION public.prune_function_logs_before(p_cutoff timestamptz)
RETURNS bigint
LANGUAGE sql
SECURITY DEFINER
SET search_path = public
AS $$
  WITH deleted AS (
    DELETE FROM function_logs
    WHERE timestamp < p_cutoff
    RETURNING 1
  )
  SELECT COUNT(*)::bigint FROM deleted;
$$;

REVOKE ALL ON FUNCTION public.prune_function_logs_before(timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.prune_function_logs_before(timestamptz) TO PUBLIC;

COMMENT ON FUNCTION public.prune_function_logs_before(timestamptz) IS
  'Deletes function_logs rows older than p_cutoff; used by orchestrator retention job.';
