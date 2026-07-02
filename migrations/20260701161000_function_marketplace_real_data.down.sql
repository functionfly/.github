DROP TRIGGER IF EXISTS trg_update_function_listing_user_rating ON function_user_ratings;
DROP FUNCTION IF EXISTS update_function_listing_user_rating();
DROP TRIGGER IF EXISTS trg_sync_function_listing_rating ON registry_function_ratings;
DROP FUNCTION IF EXISTS sync_function_listing_rating();
DROP TRIGGER IF EXISTS trg_increment_function_listing_calls ON registry_function_executions;
DROP FUNCTION IF EXISTS increment_function_listing_calls();
DROP TABLE IF EXISTS function_user_ratings;
