import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { ScrollArea, ScrollBar } from '@/components/ui/scroll-area';
import {
  ArrowLeftRight,
  Bot,
  Calendar,
  Code2,
  Database,
  FileJson,
  Globe,
  LayoutGrid,
  Lock,
  Mail,
  Sparkles,
  Webhook,
  Zap,
} from 'lucide-react';
import type { Runtime } from '../types';
import type { FunctionEditorModel } from '../useFunctionEditor';

interface Template {
  id: string;
  name: string;
  description: string;
  icon: React.ReactNode;
  runtimes: Runtime[];
  category: 'api' | 'integration' | 'scheduled' | 'utility';
  code: Record<Runtime, string>;
}

const TEMPLATES: Template[] = [
  {
    id: 'hello-world',
    name: 'Hello World',
    description: 'Basic HTTP handler with JSON response',
    icon: <Globe className="w-5 h-5" />,
    runtimes: ['typescript', 'javascript', 'python', 'go', 'deno', 'bun'],
    category: 'api',
    code: {
      typescript: `export default {
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    const url = new URL(request.url);
    
    return Response.json({
      message: 'Hello from FunctionFly!',
      path: url.pathname,
      method: request.method,
      timestamp: new Date().toISOString(),
    });
  },
};`,
      javascript: `export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    
    return new Response(
      JSON.stringify({
        message: 'Hello from FunctionFly!',
        path: url.pathname,
        method: request.method,
        timestamp: new Date().toISOString(),
      }),
      {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }
    );
  },
};`,
      python: `import json

def handler(request, env, ctx):
    """Hello World handler."""
    return {
        "status": 200,
        "headers": {"Content-Type": "application/json"},
        "body": json.dumps({
            "message": "Hello from FunctionFly!",
            "method": request.get("method", "GET"),
            "timestamp": datetime.utcnow().isoformat(),
        }),
    }`,
      go: `package main

import (
	"encoding/json"
	"net/http"
	"time"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message":   "Hello from FunctionFly!",
		"path":      r.URL.Path,
		"method":    r.Method,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}`,
      deno: `export default {
  async fetch(request: Request): Promise<Response> {
    const url = new URL(request.url);
    
    return new Response(
      JSON.stringify({
        message: 'Hello from FunctionFly!',
        path: url.pathname,
        method: request.method,
        timestamp: new Date().toISOString(),
      }),
      {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }
    );
  },
};`,
      bun: `export default {
  async fetch(request: Request): Promise<Response> {
    const url = new URL(request.url);
    
    return new Response(
      JSON.stringify({
        message: 'Hello from FunctionFly!',
        path: url.pathname,
        method: request.method,
        timestamp: new Date().toISOString(),
      }),
      {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }
    );
  },
};`,
    },
  },
  {
    id: 'crud-api',
    name: 'REST API',
    description: 'Full CRUD endpoints with in-memory store',
    icon: <Database className="w-5 h-5" />,
    runtimes: ['typescript', 'javascript', 'python', 'go'],
    category: 'api',
    code: {
      typescript: `type Item = { id: string; name: string; created: string };

const store = new Map<string, Item>();

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);
    const id = url.pathname.split('/').pop();
    
    switch (request.method) {
      case 'GET':
        if (id && id !== 'api') {
          const item = store.get(id);
          return item 
            ? Response.json(item)
            : new Response('Not found', { status: 404 });
        }
        return Response.json(Array.from(store.values()));
        
      case 'POST':
        const body = await request.json() as Omit<Item, 'id' | 'created'>;
        const newItem: Item = {
          id: crypto.randomUUID(),
          name: body.name,
          created: new Date().toISOString(),
        };
        store.set(newItem.id, newItem);
        return Response.json(newItem, { status: 201 });
        
      case 'DELETE':
        if (id && store.has(id)) {
          store.delete(id);
          return new Response(null, { status: 204 });
        }
        return new Response('Not found', { status: 404 });
        
      default:
        return new Response('Method not allowed', { status: 405 });
    }
  },
};`,
      javascript: `const store = new Map();

export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    const id = url.pathname.split('/').pop();
    
    switch (request.method) {
      case 'GET':
        if (id && id !== 'api') {
          const item = store.get(id);
          return item 
            ? Response.json(item)
            : new Response('Not found', { status: 404 });
        }
        return Response.json(Array.from(store.values()));
        
      case 'POST':
        const body = await request.json();
        const newItem = {
          id: crypto.randomUUID(),
          name: body.name,
          created: new Date().toISOString(),
        };
        store.set(newItem.id, newItem);
        return Response.json(newItem, { status: 201 });
        
      case 'DELETE':
        if (id && store.has(id)) {
          store.delete(id);
          return new Response(null, { status: 204 });
        }
        return new Response('Not found', { status: 404 });
        
      default:
        return new Response('Method not allowed', { status: 405 });
    }
  },
};`,
      python: `import json
import uuid
from datetime import datetime

store = {}

def handler(request, env, ctx):
    """CRUD API handler."""
    path = request.get("path", "/")
    method = request.get("method", "GET")
    parts = [p for p in path.split("/") if p]
    item_id = parts[-1] if parts else None
    
    if method == "GET":
        if item_id and item_id in store:
            return {"status": 200, "headers": {"Content-Type": "application/json"}, 
                    "body": json.dumps(store[item_id])}
        return {"status": 200, "headers": {"Content-Type": "application/json"},
                "body": json.dumps(list(store.values()))}
    
    elif method == "POST":
        body = json.loads(request.get("body", "{}"))
        new_item = {
            "id": str(uuid.uuid4()),
            "name": body.get("name"),
            "created": datetime.utcnow().isoformat(),
        }
        store[new_item["id"]] = new_item
        return {"status": 201, "headers": {"Content-Type": "application/json"},
                "body": json.dumps(new_item)}
    
    elif method == "DELETE":
        if item_id and item_id in store:
            del store[item_id]
            return {"status": 204, "body": ""}
        return {"status": 404, "body": "Not found"}
    
    return {"status": 405, "body": "Method not allowed"}`,
      go: `package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Item struct {
	ID      string    \`json:"id"\`
	Name    string    \`json:"name"\`
	Created time.Time \`json:"created"\`
}

var store = make(map[string]Item)

func Handler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	id := ""
	if len(parts) > 0 {
		id = parts[len(parts)-1]
	}
	
	switch r.Method {
	case "GET":
		if id != "" && id != "api" {
			if item, ok := store[id]; ok {
				json.NewEncoder(w).Encode(item)
				return
			}
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		items := make([]Item, 0, len(store))
		for _, v := range store {
			items = append(items, v)
		}
		json.NewEncoder(w).Encode(items)
		
	case "POST":
		var body struct{ Name string }
		json.NewDecoder(r.Body).Decode(&body)
		item := Item{
			ID:      uuid.New().String(),
			Name:    body.Name,
			Created: time.Now(),
		}
		store[item.ID] = item
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(item)
		
	case "DELETE":
		if _, ok := store[id]; ok {
			delete(store, id)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.Error(w, "Not found", http.StatusNotFound)
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}`,
    },
  },
  {
    id: 'webhook',
    name: 'Webhook Handler',
    description: 'Process incoming webhooks with signature validation',
    icon: <Webhook className="w-5 h-5" />,
    runtimes: ['typescript', 'javascript', 'python', 'go'],
    category: 'integration',
    code: {
      typescript: `async function verifySignature(payload: string, signature: string, secret: string): Promise<boolean> {
  const encoder = new TextEncoder();
  const data = encoder.encode(payload);
  const key = await crypto.subtle.importKey(
    'raw', encoder.encode(secret), { name: 'HMAC', hash: 'SHA-256' }, false, ['sign']
  );
  const sig = await crypto.subtle.sign('HMAC', key, data);
  const expected = 'sha256=' + btoa(String.fromCharCode(...new Uint8Array(sig)));
  return signature === expected;
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const signature = request.headers.get('X-Webhook-Signature');
    const payload = await request.text();
    
    if (!signature || !await verifySignature(payload, signature, env.WEBHOOK_SECRET)) {
      return new Response('Invalid signature', { status: 401 });
    }
    
    const event = JSON.parse(payload);
    
    // Process based on event type
    switch (event.type) {
      case 'user.created':
        // Handle new user
        console.log('New user:', event.data);
        break;
      case 'payment.success':
        // Handle payment
        console.log('Payment:', event.data);
        break;
    }
    
    return Response.json({ received: true, type: event.type });
  },
};`,
      javascript: `async function verifySignature(payload, signature, secret) {
  const encoder = new TextEncoder();
  const data = encoder.encode(payload);
  const key = await crypto.subtle.importKey(
    'raw', encoder.encode(secret), { name: 'HMAC', hash: 'SHA-256' }, false, ['sign']
  );
  const sig = await crypto.subtle.sign('HMAC', key, data);
  const expected = 'sha256=' + btoa(String.fromCharCode(...new Uint8Array(sig)));
  return signature === expected;
}

export default {
  async fetch(request, env) {
    const signature = request.headers.get('X-Webhook-Signature');
    const payload = await request.text();
    
    if (!signature || !await verifySignature(payload, signature, env.WEBHOOK_SECRET)) {
      return new Response('Invalid signature', { status: 401 });
    }
    
    const event = JSON.parse(payload);
    
    switch (event.type) {
      case 'user.created':
        console.log('New user:', event.data);
        break;
      case 'payment.success':
        console.log('Payment:', event.data);
        break;
    }
    
    return Response.json({ received: true, type: event.type });
  },
};`,
      python: `import hmac
import hashlib
import json
import base64

def verify_signature(payload, signature, secret):
    """Verify webhook signature."""
    expected = "sha256=" + base64.b64encode(
        hmac.new(secret.encode(), payload.encode(), hashlib.sha256).digest()
    ).decode()
    return hmac.compare_digest(expected, signature)

def handler(request, env, ctx):
    """Webhook handler with signature validation."""
    headers = request.get("headers", {})
    signature = headers.get("X-Webhook-Signature", "")
    payload = request.get("body", "")
    
    if not signature or not verify_signature(payload, signature, env.get("WEBHOOK_SECRET", "")):
        return {"status": 401, "body": "Invalid signature"}
    
    event = json.loads(payload)
    
    if event.get("type") == "user.created":
        print(f"New user: {event.get('data')}")
    elif event.get("type") == "payment.success":
        print(f"Payment: {event.get('data')}")
    
    return {"status": 200, "headers": {"Content-Type": "application/json"},
            "body": json.dumps({"received": True, "type": event.get("type")})}`,
      go: `package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
)

func verifySignature(payload, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	expected := "sha256=" + base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func Handler(w http.ResponseWriter, r *http.Request) {
	signature := r.Header.Get("X-Webhook-Signature")
	
	var payload map[string]interface{}
	json.NewDecoder(r.Body).Decode(&payload)
	payloadBytes, _ := json.Marshal(payload)
	
	secret := os.Getenv("WEBHOOK_SECRET")
	if signature == "" || !verifySignature(string(payloadBytes), signature, secret) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}
	
	eventType := payload["type"].(string)
	
	switch eventType {
	case "user.created":
		// Handle new user
	case "payment.success":
		// Handle payment
	}
	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"received": true,
		"type":     eventType,
	})
}`,
    },
  },
  {
    id: 'proxy',
    name: 'API Proxy',
    description: 'Forward requests with auth header injection',
    icon: <ArrowLeftRight className="w-5 h-5" />,
    runtimes: ['typescript', 'javascript', 'go'],
    category: 'utility',
    code: {
      typescript: `export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);
    
    // Construct target URL
    const targetUrl = new URL(env.TARGET_API + url.pathname);
    targetUrl.search = url.search;
    
    // Clone headers and add auth
    const headers = new Headers(request.headers);
    headers.set('Authorization', 'Bearer ' + env.API_KEY);
    headers.delete('host');
    
    // Forward request
    const response = await fetch(targetUrl.toString(), {
      method: request.method,
      headers,
      body: request.body,
    });
    
    // Return proxied response
    return new Response(response.body, {
      status: response.status,
      headers: response.headers,
    });
  },
};`,
      javascript: `export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    
    const targetUrl = new URL(env.TARGET_API + url.pathname);
    targetUrl.search = url.search;
    
    const headers = new Headers(request.headers);
    headers.set('Authorization', \`Bearer \${env.API_KEY}\`);
    headers.delete('host');
    
    const response = await fetch(targetUrl.toString(), {
      method: request.method,
      headers,
      body: request.body,
    });
    
    return new Response(response.body, {
      status: response.status,
      headers: response.headers,
    });
  },
};`,
      go: `package main

import (
	"io"
	"net/http"
	"os"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	target := os.Getenv("TARGET_API") + r.URL.Path
	
	req, _ := http.NewRequest(r.Method, target, r.Body)
	req.URL.RawQuery = r.URL.RawQuery
	
	for name, values := range r.Header {
		for _, v := range values {
			req.Header.Add(name, v)
		}
	}
	
	req.Header.Set("Authorization", "Bearer "+os.Getenv("API_KEY"))
	req.Header.Del("Host")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	
	for name, values := range resp.Header {
		for _, v := range values {
			w.Header().Add(name, v)
		}
	}
	
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}`,
    },
  },
  {
    id: 'scheduled-job',
    name: 'Scheduled Task',
    description: 'Cron job for periodic data processing',
    icon: <Calendar className="w-5 h-5" />,
    runtimes: ['typescript', 'javascript', 'python', 'go'],
    category: 'scheduled',
    code: {
      typescript: `interface ScheduledEvent {
  type: 'scheduled';
  scheduledTime: number;
}

export default {
  async fetch(request: Request): Promise<Response> {
    // For manual trigger via HTTP
    return await runTask();
  },
  
  async scheduled(event: ScheduledEvent, env: Env, ctx: ExecutionContext): Promise<void> {
    // Called by cron trigger
    console.log('Running scheduled task at', new Date(event.scheduledTime));
    await runTask();
  },
};

async function runTask(): Promise<Response> {
  try {
    // Fetch data from external API
    const response = await fetch('https://api.example.com/data');
    const data = await response.json();
    
    // Process data
    const processed = data.map((item: any) => ({
      ...item,
      processed: true,
      processedAt: new Date().toISOString(),
    }));
    
    // Store results
    await fetch(env.STORE_URL, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(processed),
    });
    
    return Response.json({ 
      success: true, 
      processed: processed.length,
      timestamp: new Date().toISOString(),
    });
  } catch (error) {
    return Response.json(
      { success: false, error: String(error) },
      { status: 500 }
    );
  }
}`,
      javascript: `export default {
  async fetch(request) {
    return await runTask();
  },
  
  async scheduled(event, env, ctx) {
    console.log('Running scheduled task at', new Date(event.scheduledTime));
    await runTask();
  },
};

async function runTask() {
  try {
    const response = await fetch('https://api.example.com/data');
    const data = await response.json();
    
    const processed = data.map((item) => ({
      ...item,
      processed: true,
      processedAt: new Date().toISOString(),
    }));
    
    await fetch(env.STORE_URL, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(processed),
    });
    
    return Response.json({ 
      success: true, 
      processed: processed.length,
      timestamp: new Date().toISOString(),
    });
  } catch (error) {
    return Response.json(
      { success: false, error: String(error) },
      { status: 500 }
    );
  }
}`,
      python: `import json
from datetime import datetime

def handler(request, env, ctx):
    """Handle both HTTP and scheduled invocations."""
    return run_task(env)

def run_task(env):
    """Run the scheduled data processing task."""
    try:
        import urllib.request
        
        # Fetch data
        req = urllib.request.Request('https://api.example.com/data')
        with urllib.request.urlopen(req) as response:
            data = json.loads(response.read().decode())
        
        # Process data
        processed = []
        for item in data:
            item['processed'] = True
            item['processed_at'] = datetime.utcnow().isoformat()
            processed.append(item)
        
        # Store results
        store_req = urllib.request.Request(
            env.get('STORE_URL', ''),
            data=json.dumps(processed).encode(),
            headers={'Content-Type': 'application/json'},
            method='POST'
        )
        urllib.request.urlopen(store_req)
        
        return {
            "status": 200,
            "headers": {"Content-Type": "application/json"},
            "body": json.dumps({
                "success": True,
                "processed": len(processed),
                "timestamp": datetime.utcnow().isoformat(),
            })
        }
    except Exception as e:
        return {
            "status": 500,
            "body": json.dumps({"success": False, "error": str(e)})
        }`,
      go: `package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"
)

type Item map[string]interface{}

func Handler(w http.ResponseWriter, r *http.Request) {
	result := runTask()
	json.NewEncoder(w).Encode(result)
}

func runTask() map[string]interface{} {
	resp, err := http.Get("https://api.example.com/data")
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	defer resp.Body.Close()
	
	var data []Item
	json.NewDecoder(resp.Body).Decode(&data)
	
	processed := make([]Item, len(data))
	for i, item := range data {
		item["processed"] = true
		item["processedAt"] = time.Now().UTC().Format(time.RFC3339)
		processed[i] = item
	}
	
	payload, _ := json.Marshal(processed)
	storeReq, _ := http.NewRequest("POST", os.Getenv("STORE_URL"), bytes.NewReader(payload))
	storeReq.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	storeResp, err := client.Do(storeReq)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	defer storeResp.Body.Close()
	
	return map[string]interface{}{
		"success":   true,
		"processed": len(processed),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
}`,
    },
  },
  {
    id: 'slack-bot',
    name: 'Slack Bot',
    description: 'Respond to Slack slash commands and events',
    icon: <Bot className="w-5 h-5" />,
    runtimes: ['typescript', 'javascript', 'python'],
    category: 'integration',
    code: {
      typescript: `async function verifySlackRequest(request: Request, signingSecret: string): Promise<boolean> {
  const timestamp = request.headers.get('X-Slack-Request-Timestamp') || '';
  const signature = request.headers.get('X-Slack-Signature') || '';
  
  // Prevent replay attacks
  const now = Math.floor(Date.now() / 1000);
  if (Math.abs(now - parseInt(timestamp)) > 300) {
    return false;
  }
  
  const body = await request.clone().text();
  const basestring = 'v0:' + timestamp + ':' + body;
  
  const encoder = new TextEncoder();
  const key = await crypto.subtle.importKey(
    'raw', encoder.encode(signingSecret), { name: 'HMAC', hash: 'SHA-256' }, false, ['sign']
  );
  const sig = await crypto.subtle.sign('HMAC', key, encoder.encode(basestring));
  const mySignature = 'v0=' + btoa(String.fromCharCode(...new Uint8Array(sig)));
  
  return signature === mySignature;
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    if (!await verifySlackRequest(request, env.SLACK_SIGNING_SECRET)) {
      return new Response('Invalid request', { status: 403 });
    }
    
    const body = await request.text();
    const params = new URLSearchParams(body);
    const command = params.get('command');
    const text = params.get('text') || '';
    
    switch (command) {
      case '/status':
        return Response.json({
          response_type: 'in_channel',
          text: '✅ System status: All services operational',
        });
        
      case '/echo':
        return Response.json({
          response_type: 'in_channel',
          text: '📢 ' + text,
        });
        
      default:
        return Response.json({
          response_type: 'ephemeral',
          text: 'Unknown command. Try /status or /echo',
        });
    }
  },
};`,
      javascript: `async function verifySlackRequest(request, signingSecret) {
  const timestamp = request.headers.get('X-Slack-Request-Timestamp') || '';
  const signature = request.headers.get('X-Slack-Signature') || '';
  
  const now = Math.floor(Date.now() / 1000);
  if (Math.abs(now - parseInt(timestamp)) > 300) {
    return false;
  }
  
  const body = await request.clone().text();
  const basestring = \`v0:\${timestamp}:\${body}\`;
  
  const encoder = new TextEncoder();
  const key = await crypto.subtle.importKey(
    'raw', encoder.encode(signingSecret), { name: 'HMAC', hash: 'SHA-256' }, false, ['sign']
  );
  const sig = await crypto.subtle.sign('HMAC', key, encoder.encode(basestring));
  const mySignature = 'v0=' + btoa(String.fromCharCode(...new Uint8Array(sig)));
  
  return signature === mySignature;
}

export default {
  async fetch(request, env) {
    if (!await verifySlackRequest(request, env.SLACK_SIGNING_SECRET)) {
      return new Response('Invalid request', { status: 403 });
    }
    
    const body = await request.text();
    const params = new URLSearchParams(body);
    const command = params.get('command');
    const text = params.get('text') || '';
    
    switch (command) {
      case '/status':
        return Response.json({
          response_type: 'in_channel',
          text: \`✅ System status: All services operational\`,
        });
        
      case '/echo':
        return Response.json({
          response_type: 'in_channel',
          text: \`📢 \${text}\`,
        });
        
      default:
        return Response.json({
          response_type: 'ephemeral',
          text: 'Unknown command. Try /status or /echo',
        });
    }
  },
};`,
      python: `import hmac
import hashlib
import base64
import json
import time
from urllib.parse import parse_qs

def verify_slack_request(request, signing_secret):
    """Verify Slack request signature."""
    headers = request.get("headers", {})
    timestamp = headers.get("X-Slack-Request-Timestamp", "")
    signature = headers.get("X-Slack-Signature", "")
    
    # Prevent replay attacks
    now = int(time.time())
    if abs(now - int(timestamp)) > 300:
        return False
    
    body = request.get("body", "")
    basestring = f"v0:{timestamp}:{body}"
    
    my_signature = "v0=" + base64.b64encode(
        hmac.new(signing_secret.encode(), basestring.encode(), hashlib.sha256).digest()
    ).decode()
    
    return hmac.compare_digest(my_signature, signature)

def handler(request, env, ctx):
    """Handle Slack slash commands."""
    if not verify_slack_request(request, env.get("SLACK_SIGNING_SECRET", "")):
        return {"status": 403, "body": "Invalid request"}
    
    body = request.get("body", "")
    params = parse_qs(body)
    command = params.get("command", [""])[0]
    text = params.get("text", [""])[0]
    
    if command == "/status":
        return {
            "status": 200,
            "headers": {"Content-Type": "application/json"},
            "body": json.dumps({
                "response_type": "in_channel",
                "text": "✅ System status: All services operational",
            })
        }
    elif command == "/echo":
        return {
            "status": 200,
            "headers": {"Content-Type": "application/json"},
            "body": json.dumps({
                "response_type": "in_channel",
                "text": f"📢 {text}",
            })
        }
    
    return {
        "status": 200,
        "headers": {"Content-Type": "application/json"},
        "body": json.dumps({
            "response_type": "ephemeral",
            "text": "Unknown command. Try /status or /echo",
        })
    }`,
    },
  },
  {
    id: 'email-sender',
    name: 'Email Sender',
    description: 'Send transactional emails via API',
    icon: <Mail className="w-5 h-5" />,
    runtimes: ['typescript', 'javascript', 'python'],
    category: 'integration',
    code: {
      typescript: `interface EmailPayload {
  to: string;
  subject: string;
  body: string;
  from?: string;
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    // Only accept POST requests
    if (request.method !== 'POST') {
      return new Response('Method not allowed', { status: 405 });
    }
    
    try {
      const payload: EmailPayload = await request.json();
      
      // Validate required fields
      if (!payload.to || !payload.subject || !payload.body) {
        return Response.json(
          { error: 'Missing required fields: to, subject, body' },
          { status: 400 }
        );
      }
      
      // Send via email service API
      const response = await fetch('https://api.resend.com/emails', {
        method: 'POST',
        headers: {
          'Authorization': 'Bearer ' + env.RESEND_API_KEY,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          from: payload.from || env.DEFAULT_FROM_EMAIL,
          to: payload.to,
          subject: payload.subject,
          html: payload.body,
        }),
      });
      
      if (!response.ok) {
        throw new Error(\`Email API error: \${response.status}\`);
      }
      
      const result = await response.json();
      
      return Response.json({
        success: true,
        messageId: result.id,
        sentAt: new Date().toISOString(),
      });
      
    } catch (error) {
      return Response.json(
        { success: false, error: String(error) },
        { status: 500 }
      );
    }
  },
};`,
      javascript: `export default {
  async fetch(request, env) {
    if (request.method !== 'POST') {
      return new Response('Method not allowed', { status: 405 });
    }
    
    try {
      const payload = await request.json();
      
      if (!payload.to || !payload.subject || !payload.body) {
        return Response.json(
          { error: 'Missing required fields: to, subject, body' },
          { status: 400 }
        );
      }
      
      const response = await fetch('https://api.resend.com/emails', {
        method: 'POST',
        headers: {
          'Authorization': \`Bearer \${env.RESEND_API_KEY}\`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          from: payload.from || env.DEFAULT_FROM_EMAIL,
          to: payload.to,
          subject: payload.subject,
          html: payload.body,
        }),
      });
      
      if (!response.ok) {
        throw new Error(\`Email API error: \${response.status}\`);
      }
      
      const result = await response.json();
      
      return Response.json({
        success: true,
        messageId: result.id,
        sentAt: new Date().toISOString(),
      });
      
    } catch (error) {
      return Response.json(
        { success: false, error: String(error) },
        { status: 500 }
      );
    }
  },
};`,
      python: `import json
import urllib.request
from datetime import datetime

def handler(request, env, ctx):
    """Send transactional emails."""
    if request.get("method") != "POST":
        return {"status": 405, "body": "Method not allowed"}
    
    try:
        payload = json.loads(request.get("body", "{}"))
        
        if not all(k in payload for k in ["to", "subject", "body"]):
            return {
                "status": 400,
                "headers": {"Content-Type": "application/json"},
                "body": json.dumps({"error": "Missing required fields: to, subject, body"})
            }
        
        req = urllib.request.Request(
            'https://api.resend.com/emails',
            data=json.dumps({
                "from": payload.get("from", env.get("DEFAULT_FROM_EMAIL", "")),
                "to": payload["to"],
                "subject": payload["subject"],
                "html": payload["body"],
            }).encode(),
            headers={
                "Authorization": f"Bearer {env.get('RESEND_API_KEY', '')}",
                "Content-Type": "application/json",
            },
            method='POST'
        )
        
        with urllib.request.urlopen(req) as response:
            result = json.loads(response.read().decode())
        
        return {
            "status": 200,
            "headers": {"Content-Type": "application/json"},
            "body": json.dumps({
                "success": True,
                "messageId": result.get("id"),
                "sentAt": datetime.utcnow().isoformat(),
            })
        }
        
    except Exception as e:
        return {
            "status": 500,
            "headers": {"Content-Type": "application/json"},
            "body": json.dumps({"success": False, "error": str(e)})
        }`,
    },
  },
  {
    id: 'jwt-auth',
    name: 'JWT Auth',
    description: 'Protected routes with JWT verification',
    icon: <Lock className="w-5 h-5" />,
    runtimes: ['typescript', 'javascript'],
    category: 'utility',
    code: {
      typescript: `// Simplified JWT verification (use a library in production)
async function verifyJWT(token: string, secret: string): Promise<{ sub?: string; exp?: number } | null> {
  const parts = token.split('.');
  if (parts.length !== 3) return null;
  
  try {
    const payload = JSON.parse(atob(parts[1]));
    
    // Check expiration
    if (payload.exp && payload.exp < Date.now() / 1000) {
      return null;
    }
    
    // Verify signature (simplified - use proper crypto in production)
    const encoder = new TextEncoder();
    const key = await crypto.subtle.importKey(
      'raw', encoder.encode(secret), { name: 'HMAC', hash: 'SHA-256' }, false, ['sign']
    );
    const sig = await crypto.subtle.sign('HMAC', key, encoder.encode(parts[0] + '.' + parts[1]));
    const expectedSig = btoa(String.fromCharCode(...new Uint8Array(sig)))
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=/g, '');
    
    if (parts[2] !== expectedSig) return null;
    
    return payload;
  } catch {
    return null;
  }
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);
    const authHeader = request.headers.get('Authorization');
    const token = authHeader?.replace('Bearer ', '');
    
    // Public health check endpoint
    if (url.pathname === '/health') {
      return Response.json({ status: 'ok' });
    }
    
    // Verify JWT for protected routes
    if (!token) {
      return Response.json(
        { error: 'Missing authorization header' },
        { status: 401 }
      );
    }
    
    const payload = await verifyJWT(token, env.JWT_SECRET);
    if (!payload) {
      return Response.json(
        { error: 'Invalid or expired token' },
        { status: 401 }
      );
    }
    
    // Return user info
    return Response.json({
      userId: payload.sub,
      authenticated: true,
      path: url.pathname,
    });
  },
};`,
      javascript: `async function verifyJWT(token, secret) {
  const parts = token.split('.');
  if (parts.length !== 3) return null;
  
  try {
    const payload = JSON.parse(atob(parts[1]));
    
    if (payload.exp && payload.exp < Date.now() / 1000) {
      return null;
    }
    
    const encoder = new TextEncoder();
    const key = await crypto.subtle.importKey(
      'raw', encoder.encode(secret), { name: 'HMAC', hash: 'SHA-256' }, false, ['sign']
    );
    const sig = await crypto.subtle.sign('HMAC', key, encoder.encode(parts[0] + '.' + parts[1]));
    const expectedSig = btoa(String.fromCharCode(...new Uint8Array(sig)))
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=/g, '');
    
    if (parts[2] !== expectedSig) return null;
    
    return payload;
  } catch {
    return null;
  }
}

export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    const authHeader = request.headers.get('Authorization');
    const token = authHeader?.replace('Bearer ', '');
    
    if (url.pathname === '/health') {
      return Response.json({ status: 'ok' });
    }
    
    if (!token) {
      return Response.json(
        { error: 'Missing authorization header' },
        { status: 401 }
      );
    }
    
    const payload = await verifyJWT(token, env.JWT_SECRET);
    if (!payload) {
      return Response.json(
        { error: 'Invalid or expired token' },
        { status: 401 }
      );
    }
    
    return Response.json({
      userId: payload.sub,
      authenticated: true,
      path: url.pathname,
    });
  },
};`,
    },
  },
  {
    id: 'json-transform',
    name: 'JSON Transform',
    description: 'Transform and filter JSON payloads',
    icon: <FileJson className="w-5 h-5" />,
    runtimes: ['typescript', 'javascript', 'python'],
    category: 'utility',
    code: {
      typescript: `interface TransformConfig {
  pick?: string[];      // Fields to keep
  omit?: string[];      // Fields to remove
  rename?: Record<string, string>;
  defaults?: Record<string, any>;
}

function transformObject(obj: any, config: TransformConfig): any {
  let result = { ...obj };
  
  // Apply defaults
  if (config.defaults) {
    result = { ...config.defaults, ...result };
  }
  
  // Pick specific fields
  if (config.pick) {
    const picked: any = {};
    for (const key of config.pick) {
      if (key in result) picked[key] = result[key];
    }
    result = picked;
  }
  
  // Omit fields
  if (config.omit) {
    for (const key of config.omit) {
      delete result[key];
    }
  }
  
  // Rename fields
  if (config.rename) {
    for (const [oldKey, newKey] of Object.entries(config.rename)) {
      if (oldKey in result) {
        result[newKey] = result[oldKey];
        delete result[oldKey];
      }
    }
  }
  
  return result;
}

export default {
  async fetch(request: Request): Promise<Response> {
    if (request.method !== 'POST') {
      return new Response('Method not allowed', { status: 405 });
    }
    
    try {
      const { data, config, batch = false } = await request.json();
      
      if (!data || !config) {
        return Response.json(
          { error: 'Missing data or config' },
          { status: 400 }
        );
      }
      
      let result;
      if (batch && Array.isArray(data)) {
        result = data.map(item => transformObject(item, config));
      } else {
        result = transformObject(data, config);
      }
      
      return Response.json({
        success: true,
        result,
        transformed: batch ? result.length : 1,
      });
      
    } catch (error) {
      return Response.json(
        { success: false, error: String(error) },
        { status: 400 }
      );
    }
  },
};`,
      javascript: `function transformObject(obj, config) {
  let result = { ...obj };
  
  if (config.defaults) {
    result = { ...config.defaults, ...result };
  }
  
  if (config.pick) {
    const picked = {};
    for (const key of config.pick) {
      if (key in result) picked[key] = result[key];
    }
    result = picked;
  }
  
  if (config.omit) {
    for (const key of config.omit) {
      delete result[key];
    }
  }
  
  if (config.rename) {
    for (const [oldKey, newKey] of Object.entries(config.rename)) {
      if (oldKey in result) {
        result[newKey] = result[oldKey];
        delete result[oldKey];
      }
    }
  }
  
  return result;
}

export default {
  async fetch(request) {
    if (request.method !== 'POST') {
      return new Response('Method not allowed', { status: 405 });
    }
    
    try {
      const { data, config, batch = false } = await request.json();
      
      if (!data || !config) {
        return Response.json(
          { error: 'Missing data or config' },
          { status: 400 }
        );
      }
      
      let result;
      if (batch && Array.isArray(data)) {
        result = data.map(item => transformObject(item, config));
      } else {
        result = transformObject(data, config);
      }
      
      return Response.json({
        success: true,
        result,
        transformed: batch ? result.length : 1,
      });
      
    } catch (error) {
      return Response.json(
        { success: false, error: String(error) },
        { status: 400 }
      );
    }
  },
};`,
      python: `import json

def transform_object(obj, config):
    """Transform an object based on config."""
    result = obj.copy()
    
    # Apply defaults
    if "defaults" in config:
        defaults = config["defaults"]
        for key, value in defaults.items():
            if key not in result:
                result[key] = value
    
    # Pick specific fields
    if "pick" in config:
        picked = {}
        for key in config["pick"]:
            if key in result:
                picked[key] = result[key]
        result = picked
    
    # Omit fields
    if "omit" in config:
        for key in config["omit"]:
            result.pop(key, None)
    
    # Rename fields
    if "rename" in config:
        for old_key, new_key in config["rename"].items():
            if old_key in result:
                result[new_key] = result.pop(old_key)
    
    return result

def handler(request, env, ctx):
    """Transform JSON payloads."""
    if request.get("method") != "POST":
        return {"status": 405, "body": "Method not allowed"}
    
    try:
        payload = json.loads(request.get("body", "{}"))
        data = payload.get("data")
        config = payload.get("config")
        batch = payload.get("batch", False)
        
        if not data or not config:
            return {
                "status": 400,
                "headers": {"Content-Type": "application/json"},
                "body": json.dumps({"error": "Missing data or config"})
            }
        
        if batch and isinstance(data, list):
            result = [transform_object(item, config) for item in data]
        else:
            result = transform_object(data, config)
        
        return {
            "status": 200,
            "headers": {"Content-Type": "application/json"},
            "body": json.dumps({
                "success": True,
                "result": result,
                "transformed": len(result) if batch else 1,
            })
        }
        
    except Exception as e:
        return {
            "status": 400,
            "headers": {"Content-Type": "application/json"},
            "body": json.dumps({"success": False, "error": str(e)})
        }`,
    },
  },
];

const CATEGORIES = [
  { id: 'all', label: 'All Templates', icon: <LayoutGrid className="w-4 h-4" /> },
  { id: 'api', label: 'APIs', icon: <Globe className="w-4 h-4" /> },
  { id: 'integration', label: 'Integrations', icon: <ArrowLeftRight className="w-4 h-4" /> },
  { id: 'scheduled', label: 'Scheduled', icon: <Calendar className="w-4 h-4" /> },
  { id: 'utility', label: 'Utilities', icon: <Zap className="w-4 h-4" /> },
];

import { Button } from '@/components/ui/button';
import { useState } from 'react';

type Props = { editor: FunctionEditorModel };

export function TemplateGallery({ editor }: Props) {
  const { runtime, setCode, handleRuntimeChange, setFunctionName, setDescription, markDirty } = editor;
  const [selectedCategory, setSelectedCategory] = useState('all');
  const [selectedTemplate, setSelectedTemplate] = useState<string | null>(null);

  const filteredTemplates = TEMPLATES.filter(
    (t) =>
      (selectedCategory === 'all' || t.category === selectedCategory) &&
      t.runtimes.includes(runtime)
  );

  const handleSelectTemplate = (template: Template) => {
    const code = template.code[runtime];
    if (code) {
      setCode(code);
      setFunctionName(template.name.toLowerCase().replace(/\s+/g, '-'));
      setDescription(template.description);
      markDirty();
      setSelectedTemplate(template.id);
    }
  };

  return (
    <Card
      className="overflow-hidden border-border-subtle/50"
      style={{ background: 'var(--bg-secondary)' }}
    >
      <CardHeader className="pb-3 pt-4 px-5">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div
              className="flex items-center justify-center w-8 h-8 rounded-lg"
              style={{
                background: 'linear-gradient(135deg, #FF6B35 0%, #FF8C42 100%)',
              }}
            >
              <Sparkles className="w-4 h-4 text-white" />
            </div>
            <div>
              <CardTitle className="text-sm font-semibold text-text-primary font-display">
                Quick Start Templates
              </CardTitle>
              <p className="text-xs text-text-muted mt-0.5">
                Choose a starter template to accelerate development
              </p>
            </div>
          </div>
          <Dialog>
            <DialogTrigger asChild>
              <Button variant="ghost" size="sm" className="h-8 text-xs gap-1.5">
                <Code2 className="w-3.5 h-3.5" />
                Browse All
              </Button>
            </DialogTrigger>
            <DialogContent
              className="sm:max-w-3xl max-h-[80vh]"
              style={{ background: 'var(--bg-secondary)' }}
            >
              <DialogHeader>
                <DialogTitle className="flex items-center gap-2 text-base font-display">
                  <LayoutGrid className="w-5 h-5 text-[#FF6B35]" />
                  Function Templates
                </DialogTitle>
              </DialogHeader>

              {/* Category filters */}
              <div className="flex flex-wrap gap-2 mt-4">
                {CATEGORIES.map((cat) => (
                  <button
                    key={cat.id}
                    onClick={() => setSelectedCategory(cat.id)}
                    className={`flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs transition-colors ${
                      selectedCategory === cat.id
                        ? 'bg-[#FF6B35] text-white'
                        : 'bg-bg-tertiary text-text-secondary hover:bg-bg-hover'
                    }`}
                  >
                    {cat.icon}
                    {cat.label}
                  </button>
                ))}
              </div>

              {/* Templates grid */}
              <ScrollArea className="h-[400px] mt-4">
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pr-4">
                  {TEMPLATES.filter(
                    (t) => selectedCategory === 'all' || t.category === selectedCategory
                  ).map((template) => (
                    <button
                      key={template.id}
                      onClick={() => handleSelectTemplate(template)}
                      className={`flex items-start gap-3 p-4 rounded-lg border-2 text-left transition-all ${
                        selectedTemplate === template.id
                          ? 'border-[#FF6B35] bg-[#FFF1EB] dark:bg-[#FF6B35]/20'
                          : 'border-border-subtle/30 bg-bg-tertiary hover:border-border-default'
                      }`}
                    >
                      <div className="p-2 rounded-lg bg-bg-secondary">{template.icon}</div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <span className="font-medium text-sm text-text-primary">
                            {template.name}
                          </span>
                          <Badge variant="outline" className="text-[10px] h-5">
                            {template.runtimes.length} runtimes
                          </Badge>
                        </div>
                        <p className="text-xs text-text-muted mt-1 leading-relaxed">
                          {template.description}
                        </p>
                        <div className="flex flex-wrap gap-1 mt-2">
                          {template.runtimes.slice(0, 4).map((r) => (
                            <span
                              key={r}
                              className="text-[10px] px-1.5 py-0.5 rounded bg-bg-secondary text-text-muted"
                            >
                              {r}
                            </span>
                          ))}
                          {template.runtimes.length > 4 && (
                            <span className="text-[10px] px-1.5 py-0.5 rounded bg-bg-secondary text-text-muted">
                              +{template.runtimes.length - 4}
                            </span>
                          )}
                        </div>
                      </div>
                    </button>
                  ))}
                </div>
                <ScrollBar orientation="vertical" />
              </ScrollArea>
            </DialogContent>
          </Dialog>
        </div>
      </CardHeader>
      <CardContent className="px-5 pb-5">
        {/* Quick template chips */}
        <ScrollArea className="w-full whitespace-nowrap">
          <div className="flex gap-2 pb-2">
            {filteredTemplates.slice(0, 6).map((template) => (
              <button
                key={template.id}
                onClick={() => handleSelectTemplate(template)}
                className={`flex items-center gap-2 px-3 py-2 rounded-lg border transition-all flex-shrink-0 ${
                  selectedTemplate === template.id
                    ? 'border-[#FF6B35] bg-[#FF6B35]/10'
                    : 'border-border-subtle/30 bg-bg-tertiary hover:border-border-default'
                }`}
              >
                <span className="text-text-muted">{template.icon}</span>
                <span className="text-xs font-medium text-text-primary whitespace-nowrap">
                  {template.name}
                </span>
              </button>
            ))}
            {filteredTemplates.length === 0 && (
              <span className="text-xs text-text-muted py-2">
                No templates available for {runtime}
              </span>
            )}
          </div>
          <ScrollBar orientation="horizontal" />
        </ScrollArea>

        {/* Info hint */}
        <p className="text-xs text-text-muted mt-3 flex items-center gap-1.5">
          <Sparkles className="w-3 h-3 text-[#FF6B35]" />
          Templates include best practices for {runtime} — customize them to fit your needs
        </p>
      </CardContent>
    </Card>
  );
}
