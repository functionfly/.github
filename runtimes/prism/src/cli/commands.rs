//! Prism CLI

use clap::{Parser, Subcommand};

/// Main CLI for Prism Runtime
#[derive(Parser, Debug)]
#[command(name = "prism")]
#[command(version = "0.1.0")]
#[command(about = "FunctionFly Prism Runtime - Universal Adaptive WASM Execution Fabric")]
pub struct Cli {
    #[command(subcommand)]
    pub command: Commands,

    /// Enable verbose logging
    #[arg(short, long)]
    pub verbose: bool,

    /// Configuration file path
    #[arg(short, long, default_value = "prism.yaml")]
    pub config: String,
}

#[derive(Subcommand, Debug)]
pub enum Commands {
    /// Start the Prism runtime
    Start {
        /// Listen address
        #[arg(short, long, default_value = "127.0.0.1:8080")]
        address: String,

        /// Enable mesh networking
        #[arg(short, long)]
        mesh: bool,
    },

    /// Execute WASM in isolated subprocess (internal use)
    Exec {
        /// WASM module path
        #[arg(long)]
        wasm: Option<String>,

        /// Cell ID for this execution
        #[arg(long)]
        cell_id: Option<String>,
    },

    /// Create a new execution cell
    Cell {
        #[command(subcommand)]
        action: CellCommands,
    },

    /// Manage capabilities
    Capability {
        #[command(subcommand)]
        action: CapabilityCommands,
    },

    /// Swarm operations
    Swarm {
        #[command(subcommand)]
        action: SwarmCommands,
    },

    /// Start interactive REPL
    Repl {
        /// Language mode (wasm, python, etc.)
        #[arg(short, long, default_value = "wasm")]
        language: String,
    },

    /// Package management
    Package {
        #[command(subcommand)]
        action: PackageCommands,
    },

    /// Show runtime status
    Status,

    /// Generate documentation
    Doc {
        /// Output directory
        #[arg(short, long, default_value = "docs")]
        output: String,
    },
}

#[derive(Subcommand, Debug)]
pub enum CellCommands {
    /// Create a new cell
    Create {
        /// WASM module path
        #[arg(short, long)]
        module: String,

        /// Memory limit in MB
        #[arg(short, long, default_value = "128")]
        memory: u64,
    },

    /// List active cells
    List,

    /// Terminate a cell
    Terminate {
        /// Cell ID
        cell_id: String,
    },

    /// Snapshot a cell
    Snapshot {
        /// Cell ID
        cell_id: String,
    },

    /// Migrate a cell
    Migrate {
        /// Cell ID
        cell_id: String,

        /// Target node ID
        #[arg(short, long)]
        target: String,
    },
}

#[derive(Subcommand, Debug)]
pub enum CapabilityCommands {
    /// Register a capability
    Register {
        /// Capability name
        name: String,

        /// Category
        #[arg(short, long)]
        category: String,
    },

    /// Discover capabilities
    Discover {
        /// Search query
        query: String,
    },

    /// List all capabilities
    List,
}

#[derive(Subcommand, Debug)]
pub enum SwarmCommands {
    /// Create a swarm
    Create {
        /// Swarm ID
        swarm_id: String,
    },

    /// Join a swarm
    Join {
        /// Swarm ID
        swarm_id: String,
    },

    /// Leave a swarm
    Leave {
        /// Swarm ID
        swarm_id: String,
    },

    /// List swarms
    List,

    /// Send command to swarm
    Command {
        /// Swarm ID
        swarm_id: String,

        /// Command
        #[arg(short, long)]
        cmd: String,
    },
}

#[derive(Subcommand, Debug)]
pub enum PackageCommands {
    /// Build a .ffpkg package
    Build {
        /// Source WASM file or directory
        #[arg(short, long)]
        source: String,

        /// Output file
        #[arg(short, long)]
        output: String,

        /// Package description
        #[arg(short, long)]
        description: Option<String>,

        /// Programming language(s) (can be specified multiple times)
        #[arg(short, long, action = clap::ArgAction::Append)]
        language: Option<Vec<String>>,

        /// Capability(s) to declare (can be specified multiple times)
        #[arg(short, long, action = clap::ArgAction::Append)]
        capability: Option<Vec<String>>,

        /// Resource file(s) to bundle (can be specified multiple times)
        #[arg(short = 'r', long, action = clap::ArgAction::Append)]
        resource: Option<Vec<String>>,
    },

    /// Inspect a package
    Inspect {
        /// Package file
        package: String,

        /// Show full details including raw JSON
        #[arg(short, long)]
        verbose: bool,
    },

    /// Sign a package
    Sign {
        /// Package file
        package: String,

        /// Private key file
        #[arg(short, long)]
        key: String,

        /// Verify signature after signing
        #[arg(short, long)]
        verify: bool,
    },
}