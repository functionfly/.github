use clap::Parser;
use tracing::{info, Level};
use tracing_subscriber::FmtSubscriber;

use sar_runtime::{
    StatefulAgentRuntime, RuntimeConfig, AgentConfig,
    Graph, Node, NodeId, NodeType, Edge,
};

#[derive(Parser, Debug)]
#[command(name = "functionfly-sar")]
#[command(about = "FunctionFly Stateful Agent Runtime")]
struct Args {
    #[arg(long, env = "NATS_URL")]
    nats_url: Option<String>,

    #[arg(long, env = "REDIS_URL")]
    redis_url: Option<String>,

    #[arg(long, env = "DATABASE_URL")]
    database_url: Option<String>,

    #[arg(long, default_value = "10000")]
    max_concurrent: usize,

    #[arg(long, default_value = "8082")]
    port: u16,
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let _subscriber = FmtSubscriber::builder()
        .with_max_level(Level::INFO)
        .with_target(true)
        .with_thread_ids(true)
        .json()
        .init();

    let args = Args::parse();

    info!(version = env!("CARGO_PKG_VERSION"), "Starting FunctionFly SAR");

    let mut config = RuntimeConfig::default();

    if let Some(url) = args.nats_url {
        config.nats_url = Some(url);
    }
    if let Some(url) = args.redis_url {
        config.redis_url = Some(url);
    }
    if let Some(url) = args.database_url {
        config.postgres_url = Some(url);
    }
    config.scheduler.max_concurrent = args.max_concurrent;

    let runtime = StatefulAgentRuntime::with_config(config).await?;

    let agent = create_sample_agent();
    runtime.register_agent(agent).await?;

    info!(port = args.port, "SAR runtime initialized");

    tokio::signal::ctrl_c().await?;

    info!("SAR runtime shutdown complete");
    Ok(())
}

fn create_sample_agent() -> AgentConfig {
    let graph_id = uuid::Uuid::new_v4();
    let input_node = NodeId(uuid::Uuid::new_v4());
    let llm_node = NodeId(uuid::Uuid::new_v4());
    let output_node = NodeId(uuid::Uuid::new_v4());

    let mut graph = Graph::new(graph_id, "sample-agent".to_string());

    graph.add_node(Node::new(input_node, "Input".to_string(), NodeType::Passthrough));
    graph.add_node(Node::new(llm_node, "LLM".to_string(), NodeType::llm("Process input".to_string())));
    graph.add_node(Node::new(output_node, "Output".to_string(), NodeType::Passthrough));

    graph.add_edge(Edge::dataflow(input_node, llm_node));
    graph.add_edge(Edge::dataflow(llm_node, output_node));

    AgentConfig::new("Sample Agent".to_string(), graph)
}
