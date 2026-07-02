-- Extend bundle_subscriptions with deployment tracking
ALTER TABLE bundle_subscriptions ADD COLUMN IF NOT EXISTS deploy_status TEXT NOT NULL DEFAULT 'pending';
ALTER TABLE bundle_subscriptions ADD COLUMN IF NOT EXISTS deploy_attempts INT NOT NULL DEFAULT 0;
ALTER TABLE bundle_subscriptions ADD COLUMN IF NOT EXISTS deploy_error TEXT;
ALTER TABLE bundle_subscriptions ADD COLUMN IF NOT EXISTS deployed_at TIMESTAMPTZ;
ALTER TABLE bundle_subscriptions ADD COLUMN IF NOT EXISTS provider_id VARCHAR(255) REFERENCES providers(id);
ALTER TABLE bundle_subscriptions ADD COLUMN IF NOT EXISTS script_name TEXT;
ALTER TABLE bundle_subscriptions ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMPTZ;

-- Index for retry ticker
CREATE INDEX IF NOT EXISTS idx_bundle_subscriptions_deploy_status
    ON bundle_subscriptions(deploy_status) WHERE deploy_status IN ('failed', 'awaiting_provider');

-- Bundle function templates
CREATE TABLE IF NOT EXISTS bundle_function_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bundle_slug TEXT NOT NULL,
    function_name TEXT NOT NULL,
    runtime TEXT NOT NULL DEFAULT 'js',
    code TEXT NOT NULL,
    route_path TEXT NOT NULL,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(bundle_slug, function_name, version)
);

CREATE INDEX IF NOT EXISTS idx_bundle_function_templates_slug
    ON bundle_function_templates(bundle_slug);

-- Seed templates from hardcoded Go code (bundles_provisioning.go:getBundleFunctionTemplates)
INSERT INTO bundle_function_templates (bundle_slug, function_name, runtime, code, route_path, version)
VALUES
    ('saas-starter', 'stripe-webhook', 'js',
'export default async (req, env) => {
  const event = await req.json();
  try {
    switch (event.type) {
      case ''customer.subscription.created'':
        await env.STATE.put(''subscriptions/'' + event.data.object.customer, JSON.stringify({
          status: ''active'',
          plan: event.data.object.items.data[0].plan.id
        }));
        break;
      case ''invoice.payment_succeeded'':
        await env.STATE.put(''payments/'' + event.data.object.id, JSON.stringify({
          status: ''paid'',
          amount: event.data.object.amount_paid
        }));
        break;
      case ''invoice.payment_failed'':
        await env.STATE.put(''failed_payments/'' + event.data.object.customer, JSON.stringify({
          timestamp: Date.now()
        }));
        break;
    }
    return new Response(JSON.stringify({ received: true }), { headers: { ''Content-Type'': ''application/json'' } });
  } catch (err) {
    return new Response(JSON.stringify({ error: err.message }), { status: 400, headers: { ''Content-Type'': ''application/json'' } });
  }
}',
'/stripe-webhook', 1),

    ('saas-starter', 'welcome-email', 'js',
'export default async (req, env) => {
  const { email: recipientEmail, name } = await req.json();
  const ffResp = await fetch(''https://api.functionfly.io/v1/email/send'', {
    method: ''POST'',
    headers: { ''Content-Type'': ''application/json'', ''Authorization'': ''Bearer '' + env.FFLY_API_KEY },
    body: JSON.stringify({ to: recipientEmail, subject: ''Welcome!'', template: ''welcome'', data: { name, email: recipientEmail } })
  });
  return new Response(JSON.stringify({ sent: ffResp.ok }), { headers: { ''Content-Type'': ''application/json'' } });
}',
'/welcome-email', 1),

    ('marketplace', 'create-listing', 'js',
'export default async (req, env) => {
  const { title, description, price } = await req.json();
  const listing = {
    id: crypto.randomUUID(),
    title,
    description,
    price_cents: Math.round(price * 100),
    status: ''active'',
    created_at: new Date().toISOString()
  };
  await env.STATE.put(''listings/'' + listing.id, JSON.stringify(listing));
  return new Response(JSON.stringify({ success: true, listing_id: listing.id }), { headers: { ''Content-Type'': ''application/json'' } });
}',
'/create-listing', 1),

    ('marketplace', 'send-message', 'js',
'export default async (req, env) => {
  const { recipient_id, content } = await req.json();
  const message = {
    id: crypto.randomUUID(),
    recipient_id,
    content,
    created_at: new Date().toISOString()
  };
  const existing = await env.STATE.get(''messages/'' + recipient_id);
  const messages = existing ? JSON.parse(existing) : [];
  messages.push(message);
  await env.STATE.put(''messages/'' + recipient_id, JSON.stringify(messages));
  return new Response(JSON.stringify({ success: true, message_id: message.id }), { headers: { ''Content-Type'': ''application/json'' } });
}',
'/send-message', 1),

    ('ai-app', 'chat-completion', 'js',
'export default async (req, env) => {
  const { message, model = ''gpt-4'' } = await req.json();
  if (!message) {
    return new Response(JSON.stringify({ error: ''message is required'' }), { status: 400, headers: { ''Content-Type'': ''application/json'' } });
  }
  const ffResp = await fetch(''https://api.functionfly.io/v1/ai/chat/completions'', {
    method: ''POST'',
    headers: { ''Content-Type'': ''application/json'', ''Authorization'': ''Bearer '' + env.FFLY_API_KEY },
    body: JSON.stringify({ model, messages: [{ role: ''user'', content: message }] })
  });
  const completion = await ffResp.json();
  return new Response(JSON.stringify({
    message: completion.choices[0].message.content,
    model: completion.model,
    usage: completion.usage
  }), { headers: { ''Content-Type'': ''application/json'' } });
}',
'/chat-completion', 1),

    ('ai-app', 'embed-and-store', 'js',
'export default async (req, env) => {
  const { content, metadata = {} } = await req.json();
  if (!content) {
    return new Response(JSON.stringify({ error: ''content is required'' }), { status: 400, headers: { ''Content-Type'': ''application/json'' } });
  }
  const ffResp = await fetch(''https://api.functionfly.io/v1/ai/embeddings'', {
    method: ''POST'',
    headers: { ''Content-Type'': ''application/json'', ''Authorization'': ''Bearer '' + env.FFLY_API_KEY },
    body: JSON.stringify({ model: ''text-embedding-3-small'', input: content })
  });
  const embedding = await ffResp.json();
  const id = crypto.randomUUID();
  await env.STATE.put(''embeddings/'' + id, JSON.stringify({
    vector: embedding.data[0].embedding,
    content: content.substring(0, 1000),
    metadata,
    created_at: new Date().toISOString()
  }));
  return new Response(JSON.stringify({ embedded: true, id }), { headers: { ''Content-Type'': ''application/json'' } });
}',
'/embed-and-store', 1)
ON CONFLICT (bundle_slug, function_name, version) DO NOTHING;
