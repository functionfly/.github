-- Restore previous validator (may reject valid code containing "ExecutionContext").

CREATE OR REPLACE FUNCTION validate_function_security(function_code text)
RETURNS boolean AS $$
BEGIN
    IF function_code ~* '(eval|exec|require\s*\(\s*.*child_process|fs\.)' THEN
        RETURN false;
    END IF;

    IF length(function_code) > 100000 THEN
        RETURN false;
    END IF;

    RETURN true;
END;
$$ LANGUAGE plpgsql IMMUTABLE;
