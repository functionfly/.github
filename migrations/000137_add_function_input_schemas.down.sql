-- Remove function input schemas table

DROP TRIGGER IF EXISTS update_function_input_schemas_updated_at ON function_input_schemas;
DROP TABLE IF EXISTS function_input_schemas;
