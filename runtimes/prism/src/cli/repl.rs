//! Interactive REPL for Prism

use std::io::{self, BufRead, Write};

/// REPL for interactive Prism sessions
pub struct Repl {
    prompt: String,
}

impl Repl {
    pub fn new() -> Self {
        Self {
            prompt: "prism> ".to_string(),
        }
    }

    /// Run the REPL
    pub fn run(&mut self) -> io::Result<()> {
        println!("\x1b[1;36mFunctionFly Prism Runtime REPL\x1b[0m");
        println!("Type 'help' for available commands, 'exit' to quit.\n");

        let stdin = io::stdin();
        let mut handle = stdin.lock();

        loop {
            print!("{}", self.prompt);
            io::stdout().flush()?;

            let mut line = String::new();
            handle.read_line(&mut line)?;

            let line = line.trim();
            if line.is_empty() {
                continue;
            }

            match line {
                "exit" | "quit" => break,
                "help" => self.print_help(),
                _ => self.execute_line(line),
            }
        }

        Ok(())
    }

    fn print_help(&self) {
        println!("Available commands:");
        println!("  help     - Show this help message");
        println!("  exit     - Exit the REPL");
        println!("  cell     - Cell operations");
        println!("  capability - Capability operations");
        println!("  swarm    - Swarm operations");
        println!("  status   - Show runtime status");
    }

    fn execute_line(&mut self, line: &str) {
        let parts: Vec<&str> = line.split_whitespace().collect();
        if parts.is_empty() {
            return;
        }

        match parts[0] {
            "cell" => {
                if parts.len() > 1 {
                    println!("Cell command: {}", parts[1..].join(" "));
                } else {
                    println!("Cell operations: create, list, status, delete");
                }
            }
            "capability" => {
                if parts.len() > 1 {
                    println!("Capability command: {}", parts[1..].join(" "));
                } else {
                    println!("Capability operations: register, discover, list");
                }
            }
            "swarm" => {
                if parts.len() > 1 {
                    println!("Swarm command: {}", parts[1..].join(" "));
                } else {
                    println!("Swarm operations: create, join, leave, list");
                }
            }
            "status" => {
                println!("Runtime status: Running");
            }
            _ => {
                println!("Unknown command: {}", line);
            }
        }
    }
}

impl Default for Repl {
    fn default() -> Self {
        Self::new()
    }
}