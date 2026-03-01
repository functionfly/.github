//! FunctionFly Node.js Runtime Binary
//! 
//! This binary provides a command-line interface for the Node.js runtime.

use std::sync::Arc;
use std::time::Duration;

use clap::{Parser, ValueEnum};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use nodejs_runtime::{
    RuntimeConfig, RuntimeVersion, create_runtime, ExecutionInput, ExecutionMetadata,
};

#[derive(Parser, Debug)]
#[command(name = "functionfly-nodejs")]
#[command(about = "FunctionFly Node.js Runtime", long_about = None)]
struct Args {
    /// Runtime version to use
    #[arg(short, long, value_enum, default_value = "node20")]
    runtime: RuntimeArg,

    /// Maximum memory in MB
    #[arg(short, long, default_value = "128")]
    memory: u32,

    /// Maximum timeout in milliseconds
    #[arg(short, long, default_value = "30000")]
    timeout: u64,

    /// Enable network access
    #[arg(long)]
    network: bool,

    /// Code to execute
    #[arg(short, long)]
    code: Option<String>,

    /// Input to pass to the function
    #[arg(short, long, default_value = "\"test\"")]
    input: String,

    /// Execute in REPL mode
    #[arg(short, long)]
    repl: bool,

    /// Verbose logging
    #[arg(short, long)]
    verbose: bool,
}

#[derive(Debug, Clone, ValueEnum)]
enum RuntimeArg {
    Node18,
    Node20,
    Deno,
}

impl From<RuntimeArg> for RuntimeVersion {
    fn from(arg: RuntimeArg) -> Self {
        match arg {
            RuntimeArg::Node18 => RuntimeVersion::Node18,
            RuntimeArg::Node20 => RuntimeVersion::Node20,
            RuntimeArg::Deno => RuntimeVersion::Deno,
        }
    }
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args = Args::parse();
    
    // Set up logging
    let log_level = if args.verbose {
        tracing::Level::DEBUG
    } else {
        tracing::Level::INFO
    };
    
    tracing_subscriber::registry()
        .with(tracing_subscriber::fmt::layer().with_level(true))
        .with(tracing_subscriber::filter::LevelFilter::from_level(log_level))
        .init();
    
    tracing::info!("FunctionFly Node.js Runtime starting...");
    tracing::info!("Runtime: {:?}", args.runtime);
    tracing::info!("Memory limit: {}MB", args.memory);
    tracing::info!("Timeout: {}ms", args.timeout);
    
    // Create runtime config
    let config = RuntimeConfig {
        version: args.runtime.into(),
        max_memory_mb: args.memory,
        max_timeout_ms: args.timeout,
        network_enabled: args.network,
        verbose_logging: args.verbose,
        ..Default::default()
    };
    
    // Validate config
    config.validate()?;
    
    // Create runtime
    let runtime = create_runtime(config)?;
    
    // Print runtime info
    let info = runtime.info();
    tracing::info!("Runtime info: {:?}", info.name);
    tracing::info!("Supported features: {:?}", info.features);
    
    // Execute code
    if let Some(code) = args.code {
        let input = ExecutionInput {
            data: serde_json::Value::String(args.input),
            metadata: ExecutionMetadata::default(),
        };
        
        tracing::info!("Executing code...");
        let result = runtime.execute(&code, input).await;
        
        if result.success {
            tracing::info!("✓ Execution successful!");
            tracing::info!("Output: {:?}", result.output);
            tracing::info!("Execution time: {}ms", result.execution_time_ms);
        } else {
            tracing::error!("✗ Execution failed: {:?}", result.error);
            std::process::exit(1);
        }
    } else if args.repl {
        tracing::info!("Starting REPL mode (not implemented)");
        // In a real implementation, this would start an interactive REPL
    } else {
        tracing::info!("No code provided. Use --code or --repl");
    }
    
    Ok(())
}
