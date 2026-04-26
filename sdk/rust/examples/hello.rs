use functionfly_sdk::{Function, Context, run};

struct HelloFunction;

impl Function for HelloFunction {
    fn handle(&self, _input: &str, _ctx: &Context) -> Result<String, String> {
        functionfly_sdk::host_functions::log("Hello from Rust function!");
        Ok(r#"{"message": "Hello from Rust!", "ok": true}"#.to_string())
    }
}

run!(HelloFunction);
