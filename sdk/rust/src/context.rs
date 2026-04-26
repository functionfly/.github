/// Execution context for a FunctionFly function.
pub struct Context {
    pub function_name: String,
    pub version: String,
}

impl Context {
    pub fn new() -> Self {
        Self {
            function_name: std::env::var("FUNCTIONFLY_FUNCTION_NAME").unwrap_or_default(),
            version: std::env::var("FUNCTIONFLY_VERSION").unwrap_or_default(),
        }
    }
}
