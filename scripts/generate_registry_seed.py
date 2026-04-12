#!/usr/bin/env python3
"""
Generate SQL INSERT statements for all functions in functions/functionfly/
Uses admin account: thefunctionfly@gmail.com
"""

import os
import json
import uuid
from datetime import datetime

FUNCTIONS_DIR = "/home/micro/projects/functionfly/functions/functionfly"
OUTPUT_SQL = "/home/micro/projects/functionfly/scripts/seed_registry_functions.sql"

def get_category_from_name(name):
    """Map function name to category"""
    categories = {
        'crypto': 'finance',
        'price': 'finance',
        'stock': 'finance',
        'finance': 'finance',
        'bank': 'finance',
        'loan': 'finance',
        'mortgage': 'finance',
        'tax': 'finance',
        'investment': 'finance',
        'budget': 'finance',
        'amortiz': 'finance',
        'annuity': 'finance',
        'geocode': 'geospatial',
        'geo': 'geospatial',
        'map': 'geospatial',
        'shapefile': 'geospatial',
        'coordinates': 'geospatial',
        'latitude': 'geospatial',
        'longitude': 'geospatial',
        'gps': 'geospatial',
        'address': 'geospatial',
        'location': 'geospatial',
        'ip': 'network',
        'dns': 'network',
        'http': 'network',
        'url': 'network',
        'request': 'network',
        'email': 'communication',
        'sms': 'communication',
        'text': 'text-processing',
        'string': 'text-processing',
        'parse': 'text-processing',
        'format': 'text-processing',
        'convert': 'text-processing',
        'classif': 'ai-ml',
        'detect': 'ai-ml',
        'recognition': 'ai-ml',
        'sentiment': 'ai-ml',
        'predict': 'ai-ml',
        'ml': 'ai-ml',
        'ai': 'ai-ml',
        'hash': 'security',
        'encrypt': 'security',
        'decrypt': 'security',
        'auth': 'security',
        'password': 'security',
        'jwt': 'security',
        'argon2': 'security',
        'aes': 'security',
        'verify': 'security',
        'validate': 'security',
        'image': 'media',
        'video': 'media',
        'audio': 'media',
        'resize': 'media',
        'compress': 'media',
        'qr': 'media',
        'barcode': 'media',
        'time': 'date-time',
        'date': 'date-time',
        'calendar': 'date-time',
        'schedule': 'date-time',
        'cron': 'date-time',
        'math': 'mathematics',
        'calc': 'mathematics',
        'stat': 'mathematics',
        'probability': 'mathematics',
        'array': 'data-structures',
        'list': 'data-structures',
        'queue': 'data-structures',
        'stack': 'data-structures',
        'tree': 'data-structures',
        'graph': 'data-structures',
        'csv': 'data-formats',
        'json': 'data-formats',
        'xml': 'data-formats',
        'yaml': 'data-formats',
        'html': 'data-formats',
        'pdf': 'data-formats',
        'qr-code': 'utilities',
        'uuid': 'utilities',
        'random': 'utilities',
        'log': 'utilities',
        'debug': 'utilities',
    }

    name_lower = name.lower()
    for keyword, category in categories.items():
        if keyword in name_lower:
            return category
    return 'utilities'

def get_tags_from_name(name):
    """Generate tags from function name"""
    tags = [name]
    parts = name.replace('-', ' ').replace('_', ' ').split()
    tags.extend(parts)
    # Add common tags based on patterns
    if 'crypto' in name.lower() or 'price' in name.lower():
        tags.extend(['finance', 'cryptocurrency', 'bitcoin'])
    if 'geocode' in name.lower() or 'geo' in name.lower():
        tags.extend(['geospatial', 'mapping', 'coordinates'])
    if 'text' in name.lower() or 'classif' in name.lower():
        tags.extend(['nlp', 'ai', 'classification'])
    return list(set(tags[:10]))  # Limit to 10 unique tags

def generate_function_inserts():
    """Generate INSERT statements for all functions"""
    functions = []

    # Get all function directories
    for func_name in sorted(os.listdir(FUNCTIONS_DIR)):
        func_path = os.path.join(FUNCTIONS_DIR, func_name)
        if not os.path.isdir(func_path):
            continue

        # Check for functionfly.jsonc or functionfly.json
        jsonc_path = os.path.join(func_path, 'functionfly.jsonc')
        json_path = os.path.join(func_path, 'functionfly.json')

        config = None
        if os.path.exists(jsonc_path):
            try:
                with open(jsonc_path) as f:
                    content = f.read()
                    # Remove comments for JSON parsing
                    lines = []
                    for line in content.split('\n'):
                        if '//' in line:
                            line = line[:line.index('//')]
                        lines.append(line)
                    config = json.loads('\n'.join(lines))
            except:
                pass
        elif os.path.exists(json_path):
            try:
                with open(json_path) as f:
                    config = json.load(f)
            except:
                pass

        # Build function data
        if config:
            title = config.get('title', func_name.replace('-', ' ').title())
            description = config.get('description', f'Function: {func_name}')
            category = config.get('category', get_category_from_name(func_name))
            version = config.get('version', '1.0.0')
            tags = config.get('tags', get_tags_from_name(func_name))
        else:
            title = func_name.replace('-', ' ').title()
            description = f'Function: {func_name}'
            category = get_category_from_name(func_name)
            version = '1.0.0'
            tags = get_tags_from_name(func_name)

        # Clean title for SQL
        title = title.replace("'", "''")
        description = description.replace("'", "''")

        functions.append({
            'name': func_name,
            'title': title,
            'description': description,
            'category': category,
            'version': version,
            'tags': tags
        })

    return functions

def generate_sql():
    """Generate complete SQL file"""
    functions = generate_function_inserts()
    print(f"Found {len(functions)} functions")

    # Fixed UUIDs for consistent seeding
    admin_user_id = "11111111-1111-1111-1111-111111111111"
    admin_tenant_id = "22222222-2222-2222-2222-222222222222"

    sql_parts = []

    # Header
    sql_parts.append("-- =============================================================================")
    sql_parts.append("-- BULK SEED: Registry Functions")
    sql_parts.append("-- Author: functionfly (thefunctionfly@gmail.com)")
    sql_parts.append(f"-- Count: {len(functions)} functions")
    sql_parts.append("-- =============================================================================")
    sql_parts.append("")

    # Create admin tenant
    sql_parts.append("-- Create admin tenant if not exists")
    sql_parts.append(f"""INSERT INTO tenants (id, name, plan, status, created_at, updated_at)
VALUES ('{admin_tenant_id}', 'functionfly-platform', 'enterprise', 'active', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET plan = 'enterprise', status = 'active';""")
    sql_parts.append("")

    # Create admin user
    sql_parts.append("-- Create admin user (thefunctionfly@gmail.com) if not exists")
    sql_parts.append(f"""INSERT INTO users (id, tenant_id, email, password_hash, role, email_verified, username, created_at, updated_at)
VALUES (
    '{admin_user_id}',
    '{admin_tenant_id}',
    'thefunctionfly@gmail.com',
    '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrHM5Oqk3VJL6K3JtFzqQJ9zqz5BLO',
    'admin',
    true,
    'functionfly',
    NOW(),
    NOW()
)
ON CONFLICT (email) DO UPDATE SET role = 'admin', tenant_id = '{admin_tenant_id}';""")
    sql_parts.append("")

    # Insert registry functions FIRST (so versions can reference them)
    sql_parts.append("-- =============================================================================")
    sql_parts.append("-- INSERT: Registry Functions (all public)")
    sql_parts.append("-- =============================================================================")
    sql_parts.append("")

    for i, func in enumerate(functions):
        func_id = str(uuid.uuid5(uuid.NAMESPACE_DNS, func['name']))
        tags_json = json.dumps(func['tags'])

        sql_parts.append(f"""INSERT INTO registry_functions (
    id, author, name, latest_version, title, description, category, tags,
    visibility, price_per_call, popularity_score, reliability_score,
    deterministic_score, tenant_id, owner_user_id, created_at, updated_at
) VALUES (
    '{func_id}'::uuid,
    'functionfly',
    '{func['name']}',
    '{func['version']}',
    '{func['title']}',
    '{func['description']}',
    '{func['category']}',
    '{tags_json}'::jsonb,
    'public',
    0.00,
    0,
    95.0,
    90.0,
    '{admin_tenant_id}'::uuid,
    '{admin_user_id}'::uuid,
    NOW(),
    NOW()
) ON CONFLICT (id) DO UPDATE SET
    visibility = 'public',
    latest_version = '{func['version']}',
    updated_at = NOW();""")

        if (i + 1) % 100 == 0:
            sql_parts.append(f"-- Completed {i + 1} functions")

    sql_parts.append("")
    sql_parts.append("-- =============================================================================")
    sql_parts.append("-- INSERT: Registry Function Versions (all active)")
    sql_parts.append("-- =============================================================================")
    sql_parts.append("")

    for i, func in enumerate(functions):
        func_id = str(uuid.uuid5(uuid.NAMESPACE_DNS, func['name']))
        version_id = str(uuid.uuid5(uuid.NAMESPACE_DNS, f"{func['name']}-version"))

        # Create minimal manifest
        manifest = {
            "version": func['version'],
            "runtime": "python3.12",
            "entry": "main.handler"
        }
        manifest_json = json.dumps(manifest).replace("'", "''")

        sql_parts.append(f"""INSERT INTO registry_function_versions (
    id, function_id, version, runtime, timeout_ms, memory_mb,
    deterministic, idempotent, cache_ttl, side_effects,
    is_active, source_path, source_code, manifest, published_at, updated_at
) VALUES (
    '{version_id}'::uuid,
    '{func_id}'::uuid,
    '{func['version']}',
    'python3.12',
    5000,
    128,
    true,
    true,
    3600,
    'none',
    true,
    'functions/functionfly/{func['name']}/main.py',
    NULL,
    '{manifest_json}'::jsonb,
    NOW(),
    NOW()
) ON CONFLICT (function_id, version) DO UPDATE SET is_active = true;""")

        if (i + 1) % 100 == 0:
            sql_parts.append(f"-- Completed {i + 1} versions")

    sql_parts.append("")
    sql_parts.append("-- =============================================================================")
    sql_parts.append("-- Verify count")
    sql_parts.append("-- =============================================================================")
    sql_parts.append("SELECT 'Total registry functions' as metric, COUNT(*) as count FROM registry_functions;")
    sql_parts.append("SELECT 'Public functions' as metric, COUNT(*) as count FROM registry_functions WHERE visibility = 'public';")
    sql_parts.append("SELECT 'Function versions' as metric, COUNT(*) as count FROM registry_function_versions;")

    return '\n'.join(sql_parts)

if __name__ == '__main__':
    sql = generate_sql()
    with open(OUTPUT_SQL, 'w') as f:
        f.write(sql)
    print(f"Generated SQL file: {OUTPUT_SQL}")
    print(f"Total functions to seed: {len(generate_function_inserts())}")
