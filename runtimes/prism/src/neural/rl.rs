//! Reinforcement learning for neural optimization

use std::collections::HashMap;
use serde::{Deserialize, Serialize};
use rand::Rng;

/// Q-Learning agent for execution optimization
#[derive(Debug, Clone)]
pub struct QLearning {
    q_table: HashMap<(String, String), f32>,
    learning_rate: f32,
    discount_factor: f32,
    epsilon: f32, // Exploration rate
}

impl QLearning {
    pub fn new(learning_rate: f32, discount_factor: f32, epsilon: f32) -> Self {
        Self {
            q_table: HashMap::new(),
            learning_rate,
            discount_factor,
            epsilon,
        }
    }

    /// Get Q-value for a state-action pair
    pub fn get_q(&self, state: &str, action: &str) -> f32 {
        self.q_table.get(&(state.to_string(), action.to_string())).copied().unwrap_or(0.0)
    }

    /// Update Q-value
    pub fn update(&mut self, state: &str, action: &str, reward: f32, next_state: &str, available_actions: &[&str]) {
        let current_q = self.get_q(state, action);

        // Max Q-value for next state
        let max_next_q = available_actions
            .iter()
            .map(|&a| self.get_q(next_state, a))
            .fold(0.0f32, |max, q| max.max(q));

        // Q-learning update rule
        let new_q = current_q + self.learning_rate * (reward + self.discount_factor * max_next_q - current_q);
        self.q_table.insert((state.to_string(), action.to_string()), new_q);
    }

    /// Select action using epsilon-greedy policy
    pub fn select_action(&self, state: &str, actions: &[&str]) -> String {
        let mut rng = rand::thread_rng();

        // Epsilon-greedy
        if rng.gen::<f32>() < self.epsilon {
            // Explore: random action
            let idx = rng.gen_range(0..actions.len());
            return actions[idx].to_string();
        }

        // Exploit: best action
        let mut best_action = actions[0].to_string();
        let mut best_q = self.get_q(state, actions[0]);

        for &action in actions {
            let q = self.get_q(state, action);
            if q > best_q {
                best_q = q;
                best_action = action.to_string();
            }
        }

        best_action
    }
}

/// Simple policy wrapper
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Policy {
    pub name: String,
    pub actions: Vec<String>,
}

impl Policy {
    pub fn new(name: &str, actions: Vec<String>) -> Self {
        Self {
            name: name.to_string(),
            actions,
        }
    }
}