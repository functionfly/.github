//! Agent Daemon Module for Always-On Agent Mode
//!
//! This module provides continuous event-driven execution for agents,
//! allowing them to react to webhooks, database changes, and scheduled triggers.

pub mod agent_daemon;
pub mod event_sources;

pub use agent_daemon::{AgentDaemon, DaemonStatus, FreeTierConfig, RunningAgent, TenantAgentState};
pub use event_sources::{WebhookEventSource, DatabaseEventSource, ScheduledEventSource};
