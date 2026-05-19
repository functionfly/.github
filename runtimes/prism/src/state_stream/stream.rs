//! State streaming for StateStream Memory Fabric

use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::{mpsc, RwLock, broadcast};

use super::store::StateSlice;
use crate::core::{StreamConfig, PrismResult};

/// A streaming handle for receiving state updates
pub struct StreamHandle {
    receiver: mpsc::Receiver<StateSlice>,
    stream_id: String,
}

/// Manages state streams for a cell
pub struct StateStream {
    config: StreamConfig,
    /// Subscribers for this stream
    subscribers: Arc<RwLock<HashMap<String, mpsc::Sender<StateSlice>>>>,
    /// Broadcast sender for fan-out
    broadcast_tx: broadcast::Sender<StateSlice>,
    /// Buffer for resumable streams
    buffer: Arc<RwLock<Vec<StateSlice>>>,
}

impl StateStream {
    pub fn new(config: StreamConfig) -> Self {
        let (broadcast_tx, _) = broadcast::channel(config.buffer_size as usize);

        Self {
            config,
            subscribers: Arc::new(RwLock::new(HashMap::new())),
            broadcast_tx,
            buffer: Arc::new(RwLock::new(Vec::new())),
        }
    }

    /// Subscribe to state updates on this stream
    pub async fn subscribe(&self, subscriber_id: &str) -> PrismResult<StreamHandle> {
        let (tx, rx) = mpsc::channel(self.config.buffer_size as usize);

        let mut subscribers = self.subscribers.write().await;
        subscribers.insert(subscriber_id.to_string(), tx);

        Ok(StreamHandle {
            receiver: rx,
            stream_id: self.config.stream_id.clone(),
        })
    }

    /// Unsubscribe from the stream
    pub async fn unsubscribe(&self, subscriber_id: &str) -> bool {
        let mut subscribers = self.subscribers.write().await;
        subscribers.remove(subscriber_id).is_some()
    }

    /// Publish a state update to all subscribers
    pub async fn publish(&self, slice: StateSlice) -> PrismResult<()> {
        // Store in buffer for resumability
        if self.config.resumable {
            let mut buffer = self.buffer.write().await;
            buffer.push(slice.clone());

            // Trim buffer if too large
            let current_len = buffer.len();
            let half_len = current_len / 2;
            if current_len > self.config.buffer_size as usize {
                buffer.drain(0..half_len);
            }
        }

        // Broadcast to all subscribers - ignore if no subscribers
        let _ = self.broadcast_tx.send(slice);

        Ok(())
    }

    /// Get buffered updates for resumption
    pub async fn get_buffer(&self) -> Vec<StateSlice> {
        let buffer = self.buffer.read().await;
        buffer.clone()
    }

    /// Clear the buffer
    pub async fn clear_buffer(&self) {
        let mut buffer = self.buffer.write().await;
        buffer.clear();
    }

    /// Get the number of active subscribers
    pub async fn subscriber_count(&self) -> usize {
        let subscribers = self.subscribers.read().await;
        subscribers.len()
    }
}

impl StreamHandle {
    /// Receive the next state slice from the stream
    pub async fn next(&mut self) -> Option<StateSlice> {
        self.receiver.recv().await
    }

    /// Get the stream ID
    pub fn stream_id(&self) -> &str {
        &self.stream_id
    }

    /// Try to receive the next state slice without waiting
    pub fn try_next(&mut self) -> Option<StateSlice> {
        self.receiver.try_recv().ok()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::core::ValueEncoding;
    use crate::state_stream::store::StateKey;

    #[tokio::test]
    async fn test_stream_subscribe() {
        let config = StreamConfig::default();
        let stream = StateStream::new(config);

        let handle = stream.subscribe("sub-1").await.unwrap();
        assert_eq!(stream.subscriber_count().await, 1);
        assert_eq!(handle.stream_id(), "");
    }

    #[tokio::test]
    async fn test_stream_publish() {
        let config = StreamConfig {
            buffer_size: 10,
            resumable: true,
            ..Default::default()
        };
        let stream = StateStream::new(config);

        let key = StateKey::new("cell-1", "counter");
        let slice = StateSlice::new(key, b"42".to_vec(), ValueEncoding::Raw);

        stream.publish(slice).await.unwrap();
        assert_eq!(stream.subscriber_count().await, 0);
    }

    #[tokio::test]
    async fn test_stream_resume() {
        let config = StreamConfig {
            buffer_size: 10,
            resumable: true,
            ..Default::default()
        };
        let stream = StateStream::new(config);

        let key = StateKey::new("cell-1", "counter");
        for i in 0..5 {
            let slice = StateSlice::new(key.clone(), format!("{}", i).into_bytes(), ValueEncoding::Raw);
            stream.publish(slice).await.unwrap();
        }

        let buffer = stream.get_buffer().await;
        assert_eq!(buffer.len(), 5);
    }
}