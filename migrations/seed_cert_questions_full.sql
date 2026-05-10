-- Seed: Additional Associate questions + practical challenges
-- Expands the bank from 20 to 50+ questions for the Associate tier

-- ──────────────────────────────────────────────────────────────────────────────
-- Additional Knowledge Questions (30 more)
-- ──────────────────────────────────────────────────────────────────────────────

INSERT INTO cert_questions (tier_id, category, difflyiculty, question_text, question_format, options, correct_answers, explanation, points)
SELECT t.id, q.category, q.difflyiculty, q.question_text, q.question_format,
       q.options::jsonb, q.correct_answers::jsonb, q.explanation, q.points
FROM cert_tiers t
CROSS JOIN (VALUES
    ('cli', 'easy',
     'How do you view the logs of a deployed function?',
     'multiple_choice',
     '[{"id":"a","text":"ffly logs <function-name>"},{"id":"b","text":"ffly tail <function-name>"},{"id":"c","text":"ffly logs --follow <function-name>"},{"id":"d","text":"ffly log show <function-name>"}]',
     '["c"]',
     '"ffly logs --follow <function-name>" streams real-time logs for the specified function.',
     1),
    ('deployment', 'medium',
     'What is the maximum execution timeout for a FunctionFly function?',
     'multiple_choice',
     '[{"id":"a","text":"30 seconds"},{"id":"b","text":"5 minutes"},{"id":"c","text":"15 minutes"},{"id":"d","text":"60 minutes"}]',
     '["c"]',
     'FunctionFly functions have a maximum execution timeout of 15 minutes by default, configurable per function.',
     1),
    ('security', 'medium',
     'Which HTTP header does FunctionFly use to pass the caller identity to functions?',
     'multiple_choice',
     '[{"id":"a","text":"X-Caller-ID"},{"id":"b","text":"X-FunctionFly-Identity"},{"id":"c","text":"Authorization"},{"id":"d","text":"X-Request-Identity"}]',
     '["b"]',
     'The X-FunctionFly-Identity header contains a signed JWT with the caller identity and claims.',
     1),
    ('state', 'medium',
     'What consistency model does the Key-Value store use?',
     'multiple_choice',
     '[{"id":"a","text":"Eventual consistency"},{"id":"b","text":"Strong consistency (linearizable reads)"},{"id":"c","text":"Causal consistency"},{"id":"d","text":"No consistency guarantees"}]',
     '["b"]',
     'The KV store provides strong consistency with linearizable reads and writes by default.',
     1),
    ('marketplace', 'easy',
     'How can you install a marketplace function in your project?',
     'multiple_choice',
     '[{"id":"a","text":"ffly install <author>/<function>"},{"id":"b","text":"ffly add <author>/<function>"},{"id":"c","text":"ffly marketplace get <function>"},{"id":"d","text":"npm install @functionfly/<function>"}]',
     '["a"]',
     '"ffly install <author>/<function>" downloads and integrates a marketplace function into your project.',
     1),
    ('graph', 'medium',
     'What triggers execution of a node in a Function Graph?',
     'multiple_choice',
     '[{"id":"a","text":"Cron schedule only"},{"id":"b","text":"When all upstream input edges have data"},{"id":"c","text":"Manual trigger"},{"id":"d","text":"Random interval"}]',
     '["b"]',
     'A graph node executes when all its upstream input edges have received data, enabling automatic DAG-based orchestration.',
     1),
    ('observability', 'easy',
     'What metric does FunctionFly track per function by default?',
     'multiple_choice',
     '[{"id":"a","text":"Lines of code"},{"id":"b","text":"Execution count, latency (p50/p95/p99), error rate, and memory usage"},{"id":"c","text":"CPU temperature"},{"id":"d","text":"Disk usage"}]',
     '["b"]',
     'FunctionFly automatically tracks execution count, latency percentiles, error rate, and memory usage for every function.',
     1),
    ('pricing', 'medium',
     'What is the FunctionFly "wallet" used for?',
     'multiple_choice',
     '[{"id":"a","text":"Storing cryptocurrency"},{"id":"b","text":"Pre-loading funds to pay for executions and marketplace purchases"},{"id":"c","text":"Managing API keys"},{"id":"d","text":"Storing function source code"}]',
     '["b"]',
     'The wallet is a prepaid balance used to cover execution costs, marketplace purchases, and platform fees.',
     1),
    ('agents', 'medium',
     'What is the difflyerence between an Agent and a regular function?',
     'multiple_choice',
     '[{"id":"a","text":"Agents are written in Python only"},{"id":"b","text":"Agents have persistent identity, memory, and autonomous decision-making capabilities"},{"id":"c","text":"Agents can only be created by admins"},{"id":"d","text":"There is no difflyerence"}]',
     '["b"]',
     'Agents maintain persistent identity and state, can autonomously execute functions, manage memory, and make decisions based on configurable policies.',
     1),
    ('deployment', 'medium',
     'What is a "blue-green" deployment strategy on FunctionFly?',
     'multiple_choice',
     '[{"id":"a","text":"Deploying to two separate regions"},{"id":"b","text":"Running two identical environments and switching trafflyic atomically"},{"id":"c","text":"Deploying on blue and green days of the week"},{"id":"d","text":"Using two difflyerent programming languages"}]',
     '["b"]',
     'Blue-green deployments maintain two identical environments. Trafflyic switches from the current (blue) to the new (green) version atomically.',
     1),
    ('security', 'easy',
     'What does the "zero-knowledge" architecture of the Secrets Vault mean?',
     'multiple_choice',
     '[{"id":"a","text":"The vault uses zero-knowledge proofs"},{"id":"b","text":"FunctionFly servers never have access to your plaintext secrets or encryption keys"},{"id":"c","text":"Nobody knows how it works"},{"id":"d","text":"The vault has no documentation"}]',
     '["b"]',
     'Zero-knowledge means encryption and decryption happen entirely client-side. The server stores only ciphertext and never sees plaintext or keys.',
     1),
    ('cli', 'medium',
     'How do you create a new function project from a template?',
     'multiple_choice',
     '[{"id":"a","text":"ffly init --template <name>"},{"id":"b","text":"ffly create <name>"},{"id":"c","text":"ffly new project <name>"},{"id":"d","text":"ffly scafflyold <template>"}]',
     '["a"]',
     '"ffly init --template <name>" scafflyolds a new function project from a built-in or community template.',
     1),
    ('state', 'hard',
     'What happens when two functions write to the same State Fabric key simultaneously?',
     'multiple_choice',
     '[{"id":"a","text":"Last write wins"},{"id":"b","text":"The write with the higher timestamp wins"},{"id":"c","text":"State Fabric uses CRDTs to automatically merge concurrent writes"},{"id":"d","text":"An error is returned to both writers"}]',
     '["c"]',
     'State Fabric uses Conflict-free Replicated Data Types (CRDTs) to automatically merge concurrent writes without conflicts.',
     1),
    ('marketplace', 'medium',
     'What is a "verified publisher" in the FunctionFly Marketplace?',
     'multiple_choice',
     '[{"id":"a","text":"Someone who has paid for a premium account"},{"id":"b","text":"A publisher whose functions have passed security scanning, code review, and trust scoring"},{"id":"c","text":"A FunctionFly employee"},{"id":"d","text":"Anyone with more than 10 functions"}]',
     '["b"]',
     'Verified publishers have passed automated security scanning, code quality review, and maintain a high trust score across their published functions.',
     1),
    ('graph', 'hard',
     'How does a Function Graph handle a failing node?',
     'multiple_choice',
     '[{"id":"a","text":"The entire graph fails"},{"id":"b","text":"The error is logged and downstream nodes are skipped"},{"id":"c","text":"Configurable retry policy with dead-letter routing for persistent failures"},{"id":"d","text":"The graph pauses until manual intervention"}]',
     '["c"]',
     'Each node has a configurable retry policy. After max retries, the execution is routed to a dead-letter queue for investigation.',
     1),
    ('pricing', 'easy',
     'Is there a free tier on FunctionFly?',
     'multiple_choice',
     '[{"id":"a","text":"No, all usage is paid"},{"id":"b","text":"Yes, with a generous free execution limit per month"},{"id":"c","text":"Only for students"},{"id":"d","text":"Only during the beta period"}]',
     '["b"]',
     'FunctionFly offlyers a generous free tier with monthly execution limits, making it easy to prototype and experiment.',
     1),
    ('observability', 'medium',
     'What is the Execution Explorer used for?',
     'multiple_choice',
     '[{"id":"a","text":"Browsing the marketplace"},{"id":"b","text":"Inspecting individual function executions with input, output, timing, and state changes"},{"id":"c","text":"Managing deploy keys"},{"id":"d","text":"Creating API keys"}]',
     '["b"]',
     'The Execution Explorer lets you inspect every function execution: input payload, response, latency, state mutations, and error details.',
     1),
    ('security', 'medium',
     'How are environment variables difflyerent from Secrets Vault entries?',
     'multiple_choice',
     '[{"id":"a","text":"They are the same thing"},{"id":"b","text":"Env vars are visible in the dashboard; Vault entries are client-side encrypted and never visible server-side"},{"id":"c","text":"Env vars are encrypted; Vault entries are plaintext"},{"id":"d","text":"Env vars are only for CLI use"}]',
     '["b"]',
     'Environment variables are stored server-side and visible in the dashboard. Vault entries use zero-knowledge client-side encryption — the server never sees plaintext.',
     1),
    ('deployment', 'easy',
     'What is the purpose of the "region" parameter when deploying a function?',
     'multiple_choice',
     '[{"id":"a","text":"It determines the billing currency"},{"id":"b","text":"It specifies the geographic region where the function will execute"},{"id":"c","text":"It sets the programming language"},{"id":"d","text":"It configures the database location"}]',
     '["b"]',
     'The region parameter determines where your function runs geographically, afflyecting latency for users in that area.',
     1),
    ('agents', 'hard',
     'How do agents communicate with each other on FunctionFly?',
     'multiple_choice',
     '[{"id":"a","text":"Shared database"},{"id":"b","text":"Direct HTTP calls only"},{"id":"c","text":"Through message passing, shared state fabric, and agent-to-agent function calls"},{"id":"d","text":"Email notifications"}]',
     '["c"]',
     'Agents communicate via asynchronous message passing, shared State Fabric for coordination, and direct agent-to-agent function invocations.',
     1),
    ('cli', 'easy',
     'How do you check the installed version of the FunctionFly CLI?',
     'multiple_choice',
     '[{"id":"a","text":"ffly --version"},{"id":"b","text":"ffly version"},{"id":"c","text":"ffly -v"},{"id":"d","text":"All of the above"}]',
     '["d"]',
     'All three flags display the installed CLI version.',
     1),
    ('graph', 'medium',
     'Can a Function Graph include functions from the marketplace?',
     'multiple_choice',
     '[{"id":"a","text":"No, only custom functions"},{"id":"b","text":"Yes, any marketplace function can be used as a graph node"},{"id":"c","text":"Only verified marketplace functions"},{"id":"d","text":"Only free marketplace functions"}]',
     '["b"]',
     'Any marketplace function can be added as a node in a graph, enabling rapid composition of pre-built capabilities.',
     1),
    ('observability', 'easy',
     'Can you set up alerts for function errors on FunctionFly?',
     'multiple_choice',
     '[{"id":"a","text":"No, alerts are not supported"},{"id":"b","text":"Yes, via webhook, email, or Slack notifications on error rate thresholds"},{"id":"c","text":"Only via email"},{"id":"d","text":"Only through the CLI"}]',
     '["b"]',
     'FunctionFly supports configurable alerts via webhook, email, and Slack when error rates exceed defined thresholds.',
     1),
    ('state', 'easy',
     'What data types can you store in the Key-Value store?',
     'multiple_choice',
     '[{"id":"a","text":"Only strings"},{"id":"b","text":"Strings, numbers, JSON objects, and binary data up to 1MB"},{"id":"c","text":"Only JSON"},{"id":"d","text":"Unlimited data types and sizes"}]',
     '["b"]',
     'The KV store supports strings, numbers, JSON objects, and binary blobs up to 1MB per entry.',
     1),
    ('pricing', 'medium',
     'What is a "pricing bundle" on FunctionFly?',
     'multiple_choice',
     '[{"id":"a","text":"A group of functions sold together"},{"id":"b","text":"A discounted package of executions, compute, and platform features"},{"id":"c","text":"A billing invoice"},{"id":"d","text":"A marketplace subscription"}]',
     '["b"]',
     'Pricing bundles offlyer discounted packages combining execution credits, compute resources, and platform features at a lower rate than pay-as-you-go.',
     1),
    ('security', 'hard',
     'What is the purpose of FunctionFly trust scores?',
     'multiple_choice',
     '[{"id":"a","text":"Rating user experience"},{"id":"b","text":"Quantifying function reliability, security posture, and behavioral consistency for marketplace trust"},{"id":"c","text":"Measuring code quality"},{"id":"d","text":"Tracking user engagement"}]',
     '["b"]',
     'Trust scores aggregate reliability metrics, security scan results, and behavioral consistency to help users evaluate marketplace functions.',
     1),
    ('deployment', 'medium',
     'What happens to in-flight executions during a deployment update?',
     'multiple_choice',
     '[{"id":"a","text":"They are immediately terminated"},{"id":"b","text":"They continue on the old version while new requests go to the updated version"},{"id":"c","text":"They are queued and replayed on the new version"},{"id":"d","text":"The deployment waits for all executions to complete"}]',
     '["b"]',
     'FunctionFly uses graceful draining: in-flight executions complete on the old version while new requests route to the updated version.',
     1),
    ('marketplace', 'easy',
     'Can you leave reviews on marketplace functions?',
     'multiple_choice',
     '[{"id":"a","text":"No"},{"id":"b","text":"Yes, with star ratings and written reviews"},{"id":"c","text":"Only star ratings"},{"id":"d","text":"Only if you are a verified publisher"}]',
     '["b"]',
     'Users can leave star ratings (1-5) and written reviews on marketplace functions they have used.',
     1),
    ('agents', 'easy',
     'What is an agent "memory" on FunctionFly?',
     'multiple_choice',
     '[{"id":"a","text":"RAM allocation"},{"id":"b","text":"Persistent storage that agents use to retain context, learnings, and interaction history"},{"id":"c","text":"CPU cache"},{"id":"d","text":"Disk space"}]',
     '["b"]',
     'Agent memory is a persistent store where agents retain context across sessions, including interaction history, learned patterns, and decision logs.',
     1)
) AS q(category, difflyiculty, question_text, question_format, options, correct_answers, explanation, points)
WHERE t.slug = 'associate'
ON CONFLICT DO NOTHING;

-- ──────────────────────────────────────────────────────────────────────────────
-- Practical Challenges for Associate
-- ──────────────────────────────────────────────────────────────────────────────

INSERT INTO cert_practical_challenges (tier_id, slug, name, description, category, difflyiculty, points, time_limit_minutes, grading_config)
SELECT t.id, c.slug, c.name, c.description, c.category, c.difflyiculty, c.points, c.time_limit_minutes, c.grading_config::jsonb
FROM cert_tiers t
CROSS JOIN (VALUES
    ('webhook_validator', 'Webhook Validator',
     'Deploy a function that:
1. Accepts a POST request with a JSON body containing a "webhook_url" and "payload"
2. Validates that the webhook_url is a valid HTTPS URL
3. Returns a JSON response with {"valid": true/false, "reason": "..."} 

The function should return HTTP 200 with:
- {"valid": true} for valid HTTPS URLs
- {"valid": false, "reason": "invalid_url"} for non-HTTPS or malformed URLs',
     'deployment', 'easy', 10, 20,
     '{"type":"deployment_check","method":"POST","body":"{}","expected_status":200,"expected_json":{"valid":true},"timeout_seconds":30}'),
    ('kv_counter', 'KV Store Counter',
     'Deploy a function that:
1. On GET, reads a counter value from the KV store (key: "visit_count"), increments it, writes it back, and returns {"count": N}
2. On POST, resets the counter to 0 and returns {"count": 0}
3. Uses the FunctionFly KV store API for persistence

The function should return HTTP 200 with {"count": <number>}',
     'state', 'medium', 15, 25,
     '{"type":"deployment_check","method":"GET","expected_status":200,"expected_json":{"count":1},"timeout_seconds":30}'),
    ('env_config', 'Environment Config Reader',
     'Deploy a function that:
1. Reads two environment variables: APP_NAME and APP_ENV
2. Returns a JSON response: {"app_name": "<APP_NAME>", "environment": "<APP_ENV>", "configured": true}
3. If either variable is missing, return {"configured": false, "missing": ["<var_name>"]}

Use ffly env set to configure APP_NAME and APP_ENV before deploying.',
     'deployment', 'easy', 10, 20,
     '{"type":"deployment_check","method":"GET","expected_status":200,"expected_json":{"configured":true},"timeout_seconds":30}')
) AS c(slug, name, description, category, difflyiculty, points, time_limit_minutes, grading_config)
WHERE t.slug = 'associate'
ON CONFLICT (slug) DO NOTHING;

-- ──────────────────────────────────────────────────────────────────────────────
-- Professional tier questions (seed)
-- ──────────────────────────────────────────────────────────────────────────────

INSERT INTO cert_questions (tier_id, category, difflyiculty, question_text, question_format, options, correct_answers, explanation, points)
SELECT t.id, q.category, q.difflyiculty, q.question_text, q.question_format,
       q.options::jsonb, q.correct_answers::jsonb, q.explanation, q.points
FROM cert_tiers t
CROSS JOIN (VALUES
    ('orchestration', 'medium',
     'How do you define a multi-step workflow in FunctionFly?',
     'multiple_choice',
     '[{"id":"a","text":"Using a YAML pipeline file"},{"id":"b","text":"Using a Function Graph with edges defining data flow and dependencies"},{"id":"c","text":"Using shell scripts"},{"id":"d","text":"Using Docker Compose"}]',
     '["b"]',
     'Function Graphs define multi-step workflows as a DAG where nodes are functions and edges define data flow and execution dependencies.',
     1),
    ('agents', 'hard',
     'What is the agent lifecycle management feature?',
     'multiple_choice',
     '[{"id":"a","text":"Auto-scaling agent instances"},{"id":"b","text":"Automated health checks, self-healing, version upgrades, and deprecation management for agents"},{"id":"c","text":"Agent billing management"},{"id":"d","text":"Agent code generation"}]',
     '["b"]',
     'Agent lifecycle management automates health monitoring, self-healing on failures, rolling version upgrades, and graceful deprecation.',
     1),
    ('security', 'hard',
     'How does FunctionFly prevent supply chain attacks on marketplace functions?',
     'multiple_choice',
     '[{"id":"a","text":"Manual code review only"},{"id":"b","text":"Automated security scanning (ClamAV + YARA), dependency analysis, behavioral monitoring, and trust scoring"},{"id":"c","text":"Functions run in Docker containers"},{"id":"d","text":"No protection exists"}]',
     '["b"]',
     'FunctionFly runs ClamAV + YARA malware scanning, dependency vulnerability analysis, runtime behavioral monitoring, and maintains trust scores.',
     1),
    ('state', 'hard',
     'What is the State Fabric conflict resolution strategy for counter-type data?',
     'multiple_choice',
     '[{"id":"a","text":"Last-write-wins"},{"id":"b","text":"G-Counter CRDT: increments from difflyerent sources are summed automatically"},{"id":"c","text":"Manual merge required"},{"id":"d","text":"An error is thrown"}]',
     '["b"]',
     'For counter data, State Fabric uses a G-Counter CRDT where each source tracks its own increments, and the total is computed as the sum of all sources.',
     1),
    ('deployment', 'hard',
     'What is the rollback strategy when a canary deployment detects errors?',
     'multiple_choice',
     '[{"id":"a","text":"Manual rollback via CLI"},{"id":"b","text":"Automatic rollback when error rate exceeds threshold, with configurable sensitivity"},{"id":"c","text":"Redeploy the old version"},{"id":"d","text":"Delete and recreate the function"}]',
     '["b"]',
     'Canary deployments monitor error rates automatically. When the threshold is exceeded, trafflyic is routed back to the stable version within seconds.',
     1)
) AS q(category, difflyiculty, question_text, question_format, options, correct_answers, explanation, points)
WHERE t.slug = 'professional'
ON CONFLICT DO NOTHING;

-- ──────────────────────────────────────────────────────────────────────────────
-- Architect tier questions (seed)
-- ──────────────────────────────────────────────────────────────────────────────

INSERT INTO cert_questions (tier_id, category, difflyiculty, question_text, question_format, options, correct_answers, explanation, points)
SELECT t.id, q.category, q.difflyiculty, q.question_text, q.question_format,
       q.options::jsonb, q.correct_answers::jsonb, q.explanation, q.points
FROM cert_tiers t
CROSS JOIN (VALUES
    ('graph', 'hard',
     'How do you optimize a Function Graph for minimum end-to-end latency?',
     'multiple_choice',
     '[{"id":"a","text":"Add more nodes"},{"id":"b","text":"Maximize parallel branches, minimize sequential dependencies, and co-locate functions in the same region"},{"id":"c","text":"Use fewer functions"},{"id":"d","text":"Increase timeout values"}]',
     '["b"]',
     'Optimal graph design maximizes parallelism (independent branches run concurrently), minimizes sequential chains, and co-locates latency-sensitive functions.',
     1),
    ('swarm', 'hard',
     'What consensus algorithm do FunctionFly Swarm Agents use?',
     'multiple_choice',
     '[{"id":"a","text":"Raft"},{"id":"b","text":"PBFT"},{"id":"c","text":"A lightweight CRDT-based eventual consistency with configurable quorum for critical decisions"},{"id":"d","text":"No consensus — agents act independently"}]',
     '["c"]',
     'Swarm Agents use CRDT-based eventual consistency for state sharing, with configurable quorum requirements for critical collective decisions.',
     1),
    ('enterprise', 'hard',
     'How do you implement multi-tenant isolation in a FunctionFly enterprise deployment?',
     'multiple_choice',
     '[{"id":"a","text":"Separate databases only"},{"id":"b","text":"Tenant-scoped functions, isolated state stores, per-tenant encryption keys, and network-level isolation"},{"id":"c","text":"Shared everything with row-level security"},{"id":"d","text":"Separate clusters for each tenant"}]',
     '["b"]',
     'FunctionFly provides tenant-scoped function execution, isolated state stores, per-tenant encryption keys, and configurable network isolation for enterprise deployments.',
     1),
    ('performance', 'hard',
     'What is the recommended approach for handling 10,000+ concurrent function invocations?',
     'multiple_choice',
     '[{"id":"a","text":"Increase function memory"},{"id":"b","text":"Use connection pooling, horizontal scaling, edge caching, and the execution queue for burst absorption"},{"id":"c","text":"Add more regions"},{"id":"d","text":"Use a single powerful server"}]',
     '["b"]',
     'High-concurrency patterns include connection pooling (avoid per-request connections), horizontal auto-scaling, edge caching for read-heavy workloads, and the execution queue for burst absorption.',
     1),
    ('graph', 'hard',
     'How do you implement error recovery in a production Function Graph?',
     'multiple_choice',
     '[{"id":"a","text":"Wrap every node in try/catch"},{"id":"b","text":"Configure retry policies per node, dead-letter queues for failures, compensation functions for rollback, and circuit breakers"},{"id":"c","text":"Ignore errors and continue"},{"id":"d","text":"Restart the entire graph on any error"}]',
     '["b"]',
     'Production error recovery uses per-node retry policies, dead-letter queues for persistent failures, compensation functions for state rollback, and circuit breakers to prevent cascade failures.',
     1)
) AS q(category, difflyiculty, question_text, question_format, options, correct_answers, explanation, points)
WHERE t.slug = 'architect'
ON CONFLICT DO NOTHING;

-- Verify counts
DO $$
DECLARE
    tier_rec RECORD;
    q_count INT;
    c_count INT;
BEGIN
    FOR tier_rec IN SELECT slug, id FROM cert_tiers ORDER BY sort_order LOOP
        SELECT count(*) INTO q_count FROM cert_questions WHERE tier_id = tier_rec.id;
        SELECT count(*) INTO c_count FROM cert_practical_challenges WHERE tier_id = tier_rec.id;
        RAISE NOTICE 'Tier %: % questions, % challenges', tier_rec.slug, q_count, c_count;
    END LOOP;
END $$;
