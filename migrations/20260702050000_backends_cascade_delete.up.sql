ALTER TABLE health_checks
    DROP CONSTRAINT health_checks_backend_id_fkey,
    ADD CONSTRAINT health_checks_backend_id_fkey
        FOREIGN KEY (backend_id) REFERENCES backends(id) ON DELETE CASCADE;

ALTER TABLE circuit_state
    DROP CONSTRAINT circuit_state_backend_id_fkey,
    ADD CONSTRAINT circuit_state_backend_id_fkey
        FOREIGN KEY (backend_id) REFERENCES backends(id) ON DELETE CASCADE;

ALTER TABLE routing_events
    DROP CONSTRAINT routing_events_backend_id_fkey,
    ADD CONSTRAINT routing_events_backend_id_fkey
        FOREIGN KEY (backend_id) REFERENCES backends(id) ON DELETE CASCADE;

ALTER TABLE registry_function_versions
    DROP CONSTRAINT registry_function_versions_backend_id_fkey,
    ADD CONSTRAINT registry_function_versions_backend_id_fkey
        FOREIGN KEY (backend_id) REFERENCES backends(id) ON DELETE CASCADE;
