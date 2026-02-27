#!/usr/bin/env python3
"""
Production-ready YARA scanning service for function verification.
Loads and compiles YARA rules from files for efficient malware detection.
"""

import os
import json
import tempfile
import hashlib
import logging
import time
from datetime import datetime
from flask import Flask, request, jsonify
from werkzeug.utils import secure_filename
import yara  # Requires python-yara package

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

app = Flask(__name__)

# Configuration
YARA_RULES_PATH = os.getenv('YARA_RULES_PATH', '/data/rules')
MAX_FILE_SIZE = 10 * 1024 * 1024  # 10MB
ALLOWED_EXTENSIONS = {'zip', 'tar', 'gz', 'js', 'py', 'ts', 'go', 'rs'}

# Global rules object - compiled once at startup
compiled_rules = None
rules_metadata = {}

def allowed_file(filename):
    return '.' in filename and \
           filename.rsplit('.', 1)[1].lower() in ALLOWED_EXTENSIONS

def load_yara_rules():
    """
    Load and compile YARA rules from files in YARA_RULES_PATH.
    Returns compiled rules object and metadata about loaded rules.
    """
    global compiled_rules, rules_metadata

    rule_files = []
    if os.path.exists(YARA_RULES_PATH):
        for file in os.listdir(YARA_RULES_PATH):
            if file.endswith(('.yar', '.yara')):
                rule_files.append(os.path.join(YARA_RULES_PATH, file))

    if not rule_files:
        logger.warning(f"No YARA rule files found in {YARA_RULES_PATH}, using fallback rule")
        # Fallback rule if no files found
        fallback_rules = '''
        rule fallback_malware {
            strings:
                $malicious = "malicious_code"
            condition:
                $malicious
        }
        '''
        try:
            compiled_rules = yara.compile(source=fallback_rules)
            rules_metadata = {
                'source': 'fallback',
                'rules_count': 1,
                'last_loaded': datetime.utcnow().isoformat() + 'Z'
            }
            logger.info("Loaded fallback YARA rule")
            return True
        except Exception as e:
            logger.error(f"Failed to compile fallback rules: {e}")
            return False

    # Load rules from files
    rules_dict = {}
    total_rules = 0

    for rule_file in rule_files:
        try:
            logger.info(f"Loading rules from {rule_file}")
            rules_dict[os.path.basename(rule_file)] = rule_file
        except Exception as e:
            logger.error(f"Failed to load {rule_file}: {e}")
            continue

    if not rules_dict:
        logger.error("No valid rule files could be loaded")
        return False

    try:
        compiled_rules = yara.compile(filepaths=rules_dict)
        rules_metadata = {
            'source': 'files',
            'rule_files': list(rules_dict.keys()),
            'rules_count': len(rules_dict),
            'last_loaded': datetime.utcnow().isoformat() + 'Z'
        }
        logger.info(f"Successfully compiled {len(rules_dict)} rule files")
        return True
    except Exception as e:
        logger.error(f"Failed to compile rules: {e}")
        return False

@app.route('/health', methods=['GET'])
def health_check():
    rules_status = "loaded" if compiled_rules is not None else "not_loaded"
    return jsonify({
        "status": "healthy" if compiled_rules is not None else "unhealthy",
        "service": "yara-scanner",
        "rules_status": rules_status,
        "rules_count": rules_metadata.get('rules_count', 0),
        "last_loaded": rules_metadata.get('last_loaded', 'never')
    })

@app.route('/scan', methods=['POST'])
def scan_file():
    if compiled_rules is None:
        return jsonify({"error": "YARA rules not loaded"}), 500

    if 'file' not in request.files:
        return jsonify({"error": "No file provided"}), 400

    file = request.files['file']
    if file.filename == '':
        return jsonify({"error": "No file selected"}), 400

    if not allowed_file(file.filename):
        return jsonify({"error": "File type not allowed"}), 400

    # Read file content
    file_content = file.read()
    if len(file_content) > MAX_FILE_SIZE:
        return jsonify({"error": "File too large"}), 413

    # Calculate file hash for caching/identification
    file_hash = hashlib.sha256(file_content).hexdigest()

    # Save to temporary file for YARA scanning
    with tempfile.NamedTemporaryFile(delete=False) as temp_file:
        temp_file.write(file_content)
        temp_file_path = temp_file.name

    start_time = time.time()
    try:
        logger.info(f"Scanning file {file.filename} (hash: {file_hash[:8]}...)")

        # Scan the file with pre-compiled rules
        matches = compiled_rules.match(temp_file_path)

        scan_time = time.time() - start_time

        result = {
            "file_hash": file_hash,
            "filename": secure_filename(file.filename),
            "file_size": len(file_content),
            "matches": [str(match) for match in matches],
            "match_count": len(matches),
            "clean": len(matches) == 0,
            "scan_time": datetime.utcnow().isoformat() + 'Z',
            "scan_duration_ms": round(scan_time * 1000, 2),
            "rules_version": rules_metadata.get('last_loaded', 'unknown')
        }

        logger.info(f"Scan completed for {file.filename}: {len(matches)} matches, {scan_time:.3f}s")
        return jsonify(result)

    except yara.Error as e:
        logger.error(f"YARA scan error for {file.filename}: {e}")
        return jsonify({"error": f"YARA scan failed: {str(e)}"}), 500
    except Exception as e:
        logger.error(f"Unexpected error scanning {file.filename}: {e}")
        return jsonify({"error": f"Scan failed: {str(e)}"}), 500

    finally:
        # Clean up temporary file
        if os.path.exists(temp_file_path):
            try:
                os.unlink(temp_file_path)
            except Exception as e:
                logger.warning(f"Failed to cleanup temp file {temp_file_path}: {e}")

@app.route('/rules', methods=['GET'])
def get_rules():
    """Get information about loaded YARA rules"""
    try:
        available_files = []
        if os.path.exists(YARA_RULES_PATH):
            for file in os.listdir(YARA_RULES_PATH):
                if file.endswith(('.yar', '.yara')):
                    available_files.append(file)

        return jsonify({
            "rules_metadata": rules_metadata,
            "available_rule_files": available_files,
            "rules_loaded": compiled_rules is not None
        })
    except Exception as e:
        logger.error(f"Error getting rules info: {e}")
        return jsonify({"error": str(e)}), 500

@app.route('/rules/reload', methods=['POST'])
def reload_rules():
    """Reload YARA rules from files"""
    try:
        logger.info("Reloading YARA rules...")
        if load_yara_rules():
            logger.info("Rules reloaded successfully")
            return jsonify({
                "status": "success",
                "message": "Rules reloaded successfully",
                "rules_metadata": rules_metadata
            })
        else:
            logger.error("Failed to reload rules")
            return jsonify({"error": "Failed to reload rules"}), 500
    except Exception as e:
        logger.error(f"Error reloading rules: {e}")
        return jsonify({"error": str(e)}), 500

if __name__ == '__main__':
    # Initialize YARA rules at startup
    logger.info("Initializing YARA scanning service...")
    if not load_yara_rules():
        logger.error("Failed to load YARA rules. Service will not start.")
        exit(1)

    port = int(os.getenv('PORT', 8080))
    logger.info(f"Starting YARA service on port {port}")
    app.run(host='0.0.0.0', port=port, debug=False)