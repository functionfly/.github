-- Fix audit trigger: ip_address column is INET, cast session text to inet
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS trigger AS $$
DECLARE
    old_row jsonb;
    new_row jsonb;
    action_type text;
    resource_id uuid;
    tenant_id uuid;
    ip_val text;
BEGIN
    IF TG_OP = 'INSERT' THEN
        action_type := 'create';
        old_row := NULL;
        new_row := row_to_json(NEW)::jsonb;
    ELSIF TG_OP = 'UPDATE' THEN
        action_type := 'update';
        old_row := row_to_json(OLD)::jsonb;
        new_row := row_to_json(NEW)::jsonb;
    ELSIF TG_OP = 'DELETE' THEN
        action_type := 'delete';
        old_row := row_to_json(OLD)::jsonb;
        new_row := NULL;
    END IF;

    BEGIN
        CASE TG_TABLE_NAME
            WHEN 'users', 'apps', 'subscriptions', 'invoices', 'usage_events', 'dashboard_configs', 'functions' THEN
                tenant_id := CASE
                    WHEN TG_OP = 'DELETE' THEN (old_row->>'tenant_id')::uuid
                    ELSE (new_row->>'tenant_id')::uuid
                END;
            WHEN 'backends', 'deployments' THEN
                DECLARE app_id uuid;
                BEGIN
                    app_id := CASE
                        WHEN TG_OP = 'DELETE' THEN (old_row->>'app_id')::uuid
                        ELSE (new_row->>'app_id')::uuid
                    END;
                    SELECT a.tenant_id INTO tenant_id FROM apps a WHERE a.id = app_id;
                END;
            WHEN 'function_deployments' THEN
                DECLARE function_id uuid;
                BEGIN
                    function_id := CASE
                        WHEN TG_OP = 'DELETE' THEN (old_row->>'function_id')::uuid
                        ELSE (new_row->>'function_id')::uuid
                    END;
                    SELECT f.tenant_id INTO tenant_id FROM functions f WHERE f.id = function_id;
                END;
            ELSE
                tenant_id := NULL;
        END CASE;
    EXCEPTION WHEN OTHERS THEN
        tenant_id := NULL;
    END;

    BEGIN
        resource_id := CASE
            WHEN TG_OP = 'DELETE' THEN (old_row->>'id')::uuid
            ELSE (new_row->>'id')::uuid
        END;
    EXCEPTION WHEN OTHERS THEN
        resource_id := NULL;
    END;

    ip_val := NULLIF(current_setting('app.client_ip', true), '');

    INSERT INTO audit_events (
        actor_user_id,
        tenant_id,
        action,
        resource_type,
        resource_id,
        before_state,
        after_state,
        ip_address,
        user_agent,
        timestamp,
        success
    ) VALUES (
        current_user_id(),
        tenant_id,
        action_type || '_' || TG_TABLE_NAME,
        TG_TABLE_NAME,
        resource_id,
        old_row,
        new_row,
        CASE WHEN ip_val = '' OR ip_val IS NULL THEN NULL ELSE ip_val::inet END,
        NULLIF(current_setting('app.user_agent', true), ''),
        now(),
        true
    );

    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSE
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
