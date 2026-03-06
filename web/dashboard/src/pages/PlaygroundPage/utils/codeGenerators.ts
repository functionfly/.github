/**
 * Code snippet generators for multiple languages.
 * Given an author, function name, and input value, generates ready-to-use code.
 */

import { getApiBaseUrl } from '@/lib/constants';

export type SnippetLanguage = 'curl' | 'javascript' | 'typescript' | 'python' | 'go' | 'php';

export interface SnippetOptions {
  author: string;
  name: string;
  input: unknown;
  baseUrl?: string;
  apiKey?: string;
}

function getBaseUrl(options: SnippetOptions): string {
  return options.baseUrl ?? getApiBaseUrl();
}

function getInputBody(input: unknown): string {
  return JSON.stringify(input, null, 2);
}

export function generateCurl(options: SnippetOptions): string {
  const url = `${getBaseUrl(options)}/v1/fx/${options.author}/${options.name}`;
  const body = getInputBody(options.input);
  const authHeader = options.apiKey ? `\\\n  -H "Authorization: Bearer ${options.apiKey}" ` : '';

  return `curl -X POST "${url}" \\
  -H "Content-Type: application/json" ${authHeader}\\
  -d '${body}'`;
}

export function generateJavaScript(options: SnippetOptions): string {
  const url = `${getBaseUrl(options)}/v1/fx/${options.author}/${options.name}`;
  const body = getInputBody(options.input);
  const authHeader = options.apiKey
    ? `\n    "Authorization": "Bearer ${options.apiKey}",`
    : '';

  return `const response = await fetch("${url}", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",${authHeader}
  },
  body: JSON.stringify(${body}),
});

const result = await response.json();
console.log(result);`;
}

export function generateTypeScript(options: SnippetOptions): string {
  const url = `${getBaseUrl(options)}/v1/fx/${options.author}/${options.name}`;
  const body = getInputBody(options.input);
  const authHeader = options.apiKey
    ? `\n    "Authorization": \`Bearer ${options.apiKey}\`,`
    : '';

  return `interface FunctionResult {
  ok: boolean;
  data?: unknown;
  cached: boolean;
  duration_ms: number;
  version: string;
  error?: { code: string; message: string };
}

const response = await fetch("${url}", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",${authHeader}
  },
  body: JSON.stringify(${body}),
});

const result: FunctionResult = await response.json();

if (result.ok) {
  console.log("Success:", result.data);
} else {
  console.error("Error:", result.error?.message);
}`;
}

export function generatePython(options: SnippetOptions): string {
  const url = `${getBaseUrl(options)}/v1/fx/${options.author}/${options.name}`;
  const body = getInputBody(options.input);
  const authHeader = options.apiKey
    ? `\n    "Authorization": f"Bearer ${options.apiKey}",`
    : '';

  return `import requests

url = "${url}"
headers = {
    "Content-Type": "application/json",${authHeader}
}
payload = ${body}

response = requests.post(url, json=payload, headers=headers)
result = response.json()

if result.get("ok"):
    print("Success:", result.get("data"))
else:
    print("Error:", result.get("error", {}).get("message"))`;
}

export function generateGo(options: SnippetOptions): string {
  const url = `${getBaseUrl(options)}/v1/fx/${options.author}/${options.name}`;
  const body = getInputBody(options.input);
  const authHeader = options.apiKey
    ? `\n\treq.Header.Set("Authorization", "Bearer "+apiKey)`
    : '';

  return `package main

import (
\t"bytes"
\t"encoding/json"
\t"fmt"
\t"net/http"
)

func main() {
\tpayload := ${body}

\tbody, _ := json.Marshal(payload)
\treq, _ := http.NewRequest("POST", "${url}", bytes.NewBuffer(body))
\treq.Header.Set("Content-Type", "application/json")${authHeader}

\tclient := &http.Client{}
\tresp, err := client.Do(req)
\tif err != nil {
\t\tpanic(err)
\t}
\tdefer resp.Body.Close()

\tvar result map[string]interface{}
\tjson.NewDecoder(resp.Body).Decode(&result)
\tfmt.Println(result)
}`;
}

export function generatePhp(options: SnippetOptions): string {
  const url = `${getBaseUrl(options)}/v1/fx/${options.author}/${options.name}`;
  const body = getInputBody(options.input);
  const authHeader = options.apiKey
    ? `\n    "Authorization: Bearer ${options.apiKey}",`
    : '';

  return `<?php

$url = "${url}";
$data = ${body};

$ch = curl_init($url);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_POST, true);
curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($data));
curl_setopt($ch, CURLOPT_HTTPHEADER, [
    "Content-Type: application/json",${authHeader}
]);

$response = curl_exec($ch);
curl_close($ch);

$result = json_decode($response, true);

if ($result["ok"]) {
    echo "Success: " . json_encode($result["data"]);
} else {
    echo "Error: " . $result["error"]["message"];
}`;
}

export function generateSnippet(language: SnippetLanguage, options: SnippetOptions): string {
  switch (language) {
    case 'curl':
      return generateCurl(options);
    case 'javascript':
      return generateJavaScript(options);
    case 'typescript':
      return generateTypeScript(options);
    case 'python':
      return generatePython(options);
    case 'go':
      return generateGo(options);
    case 'php':
      return generatePhp(options);
    default:
      return generateCurl(options);
  }
}

export const SNIPPET_LANGUAGES: Array<{ id: SnippetLanguage; label: string; syntaxLang: string }> = [
  { id: 'curl', label: 'cURL', syntaxLang: 'bash' },
  { id: 'javascript', label: 'JavaScript', syntaxLang: 'javascript' },
  { id: 'typescript', label: 'TypeScript', syntaxLang: 'typescript' },
  { id: 'python', label: 'Python', syntaxLang: 'python' },
  { id: 'go', label: 'Go', syntaxLang: 'go' },
  { id: 'php', label: 'PHP', syntaxLang: 'php' },
];
