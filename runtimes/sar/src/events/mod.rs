use std::collections::HashMap;
use std::sync::Arc;

use parking_lot::RwLock;
use serde::{Deserialize, Serialize};
use tracing::info;
use uuid::Uuid;

use crate::core::AgentId;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Event {
    pub id: Uuid,
    pub source: EventSource,
    pub subject: String,
    pub payload: serde_json::Value,
    pub metadata: HashMap<String, String>,
    pub timestamp: chrono::DateTime<chrono::Utc>,
}

impl Event {
    pub fn new(source: EventSource, subject: String, payload: serde_json::Value) -> Self {
        Self {
            id: Uuid::new_v4(),
            source,
            subject,
            payload,
            metadata: HashMap::new(),
            timestamp: chrono::Utc::now(),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum EventSource {
    Webhook { url: String },
    Database,
    MessageQueue,
    InternalAgent { agent_id: AgentId },
    Cron,
    Api,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EventSubscription {
    pub id: Uuid,
    pub agent_id: AgentId,
    pub subject: String,
    pub filter: Option<serde_json::Value>,
}

impl EventSubscription {
    pub fn new(agent_id: AgentId, subject: String) -> Self {
        Self {
            id: Uuid::new_v4(),
            agent_id,
            subject,
            filter: None,
        }
    }
}

#[cfg(feature = "nats-events")]
pub struct NatsEventBus {
    client: Option<nats::Connection>,
    subscriptions: Arc<RwLock<HashMap<String, Vec<EventSubscription>>>>,
}

#[cfg(feature = "nats-events")]
impl NatsEventBus {
    pub fn new(url: String) -> anyhow::Result<Self> {
        let client = nats::connect(url)?;
        Ok(Self {
            client: Some(client),
            subscriptions: Arc::new(RwLock::new(HashMap::new())),
        })
    }

    pub fn subscribe(&self, agent_id: AgentId, subject: String) -> anyhow::Result<()> {
        let sub = EventSubscription::new(agent_id, subject.clone());
        let mut subs = self.subscriptions.write();
        subs.entry(subject).or_default().push(sub);
        Ok(())
    }

    pub fn publish(&self, event: Event) -> anyhow::Result<()> {
        if let Some(ref client) = self.client {
            let payload = serde_json::to_vec(&event)?;
            client.publish(&event.subject, &payload)?;
            info!(subject = %event.subject, "Event published");
        }
        Ok(())
    }

    pub fn subscribe_nats(&self, subject: &str) -> anyhow::Result<nats::Subscription> {
        if let Some(ref client) = self.client {
            Ok(client.subscribe(subject)?)
        } else {
            Err(anyhow::anyhow!("NATS not connected"))
        }
    }
}

#[cfg(not(feature = "nats-events"))]
pub struct NatsEventBus;

#[cfg(not(feature = "nats-events"))]
impl NatsEventBus {
    pub fn new(_url: String) -> anyhow::Result<Self> {
        Ok(Self)
    }

    pub fn subscribe(&self, _agent_id: AgentId, _subject: String) -> anyhow::Result<()> {
        Ok(())
    }

    pub fn publish(&self, _event: Event) -> anyhow::Result<()> {
        Ok(())
    }
}

#[cfg(feature = "nats-events")]
impl Default for NatsEventBus {
    fn default() -> Self {
        Self {
            client: None,
            subscriptions: Arc::new(RwLock::new(HashMap::new())),
        }
    }
}

#[cfg(not(feature = "nats-events"))]
impl Default for NatsEventBus {
    fn default() -> Self {
        Self
    }
}

#[derive(Debug, Clone)]
pub struct EventProcessor {
    subscriptions: Arc<RwLock<HashMap<String, Vec<EventSubscription>>>>,
}

impl EventProcessor {
    pub fn new() -> Self {
        Self {
            subscriptions: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    pub fn subscribe(&self, agent_id: AgentId, subject: String) {
        let sub = EventSubscription::new(agent_id, subject);
        let mut subs = self.subscriptions.write();
        subs.entry(sub.subject.clone()).or_default().push(sub);
    }

    pub fn unsubscribe(&self, agent_id: AgentId, subject: &str) {
        let mut subs = self.subscriptions.write();
        if let Some(subs_list) = subs.get_mut(subject) {
            subs_list.retain(|s| s.agent_id != agent_id);
        }
    }

    pub fn get_subscriptions(&self, subject: &str) -> Vec<EventSubscription> {
        let subs = self.subscriptions.read();
        subs.get(subject).cloned().unwrap_or_default()
    }
}

impl Default for EventProcessor {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_event_creation() {
        let event = Event::new(
            EventSource::Api,
            "agent.task.completed".to_string(),
            serde_json::json!({"task_id": "123"}),
        );
        assert_eq!(event.subject, "agent.task.completed");
        assert_eq!(event.payload["task_id"], "123");
    }

    #[test]
    fn test_event_processor_subscribe() {
        let processor = EventProcessor::new();
        let agent_id = AgentId(Uuid::new_v4());

        processor.subscribe(agent_id, "webhook.stripe".to_string());

        let subs = processor.get_subscriptions("webhook.stripe");
        assert_eq!(subs.len(), 1);
        assert_eq!(subs[0].agent_id, agent_id);
    }

    #[test]
    fn test_event_processor_unsubscribe() {
        let processor = EventProcessor::new();
        let agent_id = AgentId(Uuid::new_v4());

        processor.subscribe(agent_id, "webhook.stripe".to_string());
        assert_eq!(processor.get_subscriptions("webhook.stripe").len(), 1);

        processor.unsubscribe(agent_id, "webhook.stripe");
        assert_eq!(processor.get_subscriptions("webhook.stripe").len(), 0);
    }

    #[test]
    fn test_event_processor_multiple_agents() {
        let processor = EventProcessor::new();
        let agent1 = AgentId(Uuid::new_v4());
        let agent2 = AgentId(Uuid::new_v4());

        processor.subscribe(agent1, "events.db".to_string());
        processor.subscribe(agent2, "events.db".to_string());

        let subs = processor.get_subscriptions("events.db");
        assert_eq!(subs.len(), 2);
    }
}
