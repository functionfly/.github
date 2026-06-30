"""
FunctionFly WASM Host Function Bridge

This module provides Python-callable wrappers for FunctionFly host functions
available in the WASM runtime. It uses a shared memory protocol with the
wrapper module's ff_invoke export to call host functions from Python code
running inside MicroPython.

The host function call protocol uses a fixed memory buffer:
  - Bytes 0-3:   function ID (i32 LE)
  - Bytes 4-7:   arg1_ptr (i32 LE)
  - Bytes 8-11:  arg1_len (i32 LE)
  - Bytes 12-15: arg2_ptr (i32 LE)
  - Bytes 16-19: arg2_len (i32 LE)
  - Bytes 20-23: status (i32 LE) - written by host
  - Bytes 24-27: result_buf_ptr (i32 LE)
  - Bytes 28-31: result_len_ptr (i32 LE)

Function IDs:
  1 = state_get, 2 = state_set, 3 = state_delete
  4 = state_get_fabric, 5 = state_create_snapshot
  6 = get_env, 7 = kv_get, 8 = kv_set, 9 = log
"""

import json
import struct

# Host call buffer address and size
# This matches the wrapper module's memory layout
_HOST_CALL_BUF = 0xF0200  # After dynamic_base (0xF0000) + 512 bytes padding
_HOST_CALL_BUF_SIZE = 4096

# Result buffer is after the call buffer
_RESULT_BUF = _HOST_CALL_BUF + 256
_RESULT_BUF_SIZE = _HOST_CALL_BUF_SIZE - 256

# Function IDs
_FN_STATE_GET = 1
_FN_STATE_SET = 2
_FN_STATE_DELETE = 3
_FN_STATE_GET_FABRIC = 4
_FN_STATE_CREATE_SNAPSHOT = 5
_FN_GET_ENV = 6
_FN_KV_GET = 7
_FN_KV_SET = 8
_FN_LOG = 9
_FN_GET_ATTESTATION = 10
_FN_DELEGATE = 11


def _is_wasm_environment():
    """Check if we're running in the FunctionFly WASM runtime."""
    try:
        import uctypes
        # Try to access the host call buffer signature
        buf = uctypes.bytearray_at(_HOST_CALL_BUF, 4)
        return True
    except (ImportError, OSError, AttributeError):
        return False


def _write_string_to_buf(buf, offset, data):
    """Write a string to the buffer at the given offset, returning (ptr, len)."""
    encoded = data.encode('utf-8') if isinstance(data, str) else data
    buf[offset:offset + len(encoded)] = encoded
    return offset, len(encoded)


def _read_string_from_buf(buf, offset, length):
    """Read a string from the buffer."""
    return bytes(buf[offset:offset + length]).decode('utf-8')


def _read_le32(buf, offset):
    """Read a little-endian i32 from buffer."""
    return struct.unpack('<i', buf[offset:offset + 4])[0]


def _write_le32(buf, offset, value):
    """Write a little-endian i32 to buffer."""
    struct.pack_into('<i', buf, offset, value)


def _invoke_host_fn(fn_id, arg1_str='', arg2_str=''):
    """
    Invoke a host function via the shared memory protocol.

    Args:
        fn_id: Function ID constant
        arg1_str: First string argument
        arg2_str: Second string argument (optional)

    Returns:
        tuple: (status_code, result_string)
    """
    try:
        import uctypes
    except ImportError:
        raise RuntimeError(
            "FunctionFly host functions require uctypes module (WASM runtime)"
        )

    # Get direct memory access to the host call buffer
    buf = uctypes.bytearray_at(_HOST_CALL_BUF, _HOST_CALL_BUF_SIZE)

    # Encode arguments
    arg1_bytes = arg1_str.encode('utf-8') if arg1_str else b''
    arg2_bytes = arg2_str.encode('utf-8') if arg2_str else b''

    # Write arg1 into the result buffer area (we use it as temp space for args)
    arg1_offset = _RESULT_BUF + _RESULT_BUF_SIZE // 2
    if arg1_bytes:
        buf[arg1_offset - _HOST_CALL_BUF:arg1_offset - _HOST_CALL_BUF + len(arg1_bytes)] = arg1_bytes

    # Write arg2 after arg1
    arg2_offset = arg1_offset + len(arg1_bytes)
    if arg2_bytes:
        buf[arg2_offset - _HOST_CALL_BUF:arg2_offset - _HOST_CALL_BUF + len(arg2_bytes)] = arg2_bytes

    # Build the call buffer
    _write_le32(buf, 0, fn_id)                          # fn_id
    _write_le32(buf, 4, arg1_offset if arg1_bytes else 0)  # arg1_ptr
    _write_le32(buf, 8, len(arg1_bytes))                 # arg1_len
    _write_le32(buf, 12, arg2_offset if arg2_bytes else 0) # arg2_ptr
    _write_le32(buf, 16, len(arg2_bytes))                # arg2_len
    _write_le32(buf, 20, -1)                             # status (pending)
    _write_le32(buf, 24, _RESULT_BUF)                    # result_buf_ptr
    _write_le32(buf, 28, _RESULT_BUF + _RESULT_BUF_SIZE - 4)  # result_len_ptr

    # Call the wrapper's ff_invoke export
    # In WASM, this would be a direct function call
    # For now, the host monitors the status field via mp_js_hook
    # Set status to 0 to signal "ready to process"
    _write_le32(buf, 20, 0)

    # Busy-wait for the host to process (mp_js_hook checks this)
    # The host sets status back to the result code when done
    import time
    max_wait = 1000  # iterations
    for _ in range(max_wait):
        status = _read_le32(buf, 20)
        if status != 0:
            break
        # Yield to allow mp_js_hook to run
        try:
            time.sleep(0.001)
        except (OSError, AttributeError):
            pass  # Some WASM environments may not support sleep

    status = _read_le32(buf, 20)
    if status == 0:
        return -1, ''  # Timeout

    # Read result
    result_len_ptr = _RESULT_BUF + _RESULT_BUF_SIZE - 4
    result_len = _read_le32(buf, result_len_ptr - _HOST_CALL_BUF)

    result_str = ''
    if result_len > 0 and status > 0:
        result_bytes = buf[_RESULT_BUF - _HOST_CALL_BUF:_RESULT_BUF - _HOST_CALL_BUF + min(result_len, _RESULT_BUF_SIZE)]
        try:
            result_str = result_bytes.decode('utf-8')
        except UnicodeDecodeError:
            result_str = ''

    return status, result_str


def log(level, message):
    """Log a message via the FunctionFly logging system.

    Args:
        level: Log level (0=debug, 1=info, 2=warn, 3=error)
        message: Log message string
    """
    try:
        _invoke_host_fn(_FN_LOG, str(level), message)
    except Exception:
        pass  # Logging should never fail


def get_env(name):
    """Get an environment variable.

    Args:
        name: Variable name

    Returns:
        Variable value, or None if not found
    """
    status, result = _invoke_host_fn(_FN_GET_ENV, name)
    if status == 0:
        return result
    return None


def kv_get(key):
    """Get a value from the key-value store.

    Args:
        key: The key to look up

    Returns:
        The value, or None if not found
    """
    status, result = _invoke_host_fn(_FN_KV_GET, key)
    if status == 0:
        return result
    return None


def kv_set(key, value):
    """Set a value in the key-value store.

    Args:
        key: The key
        value: The value to store

    Returns:
        True if successful
    """
    status, _ = _invoke_host_fn(_FN_KV_SET, key, value)
    return status == 0


def state_get(path):
    """Get a value from StateFabric.

    Args:
        path: Full state path (tenant/fabric/key)

    Returns:
        JSON string with value and metadata, or None on error
    """
    status, result = _invoke_host_fn(_FN_STATE_GET, path)
    if status == 0:
        return result
    return None


def state_set(path, value):
    """Set a value in StateFabric.

    Args:
        path: Full state path (tenant/fabric/key)
        value: JSON string of the value to store

    Returns:
        True if successful
    """
    status, _ = _invoke_host_fn(_FN_STATE_SET, path, value)
    return status == 0


def state_delete(path):
    """Delete a value from StateFabric.

    Args:
        path: Full state path (tenant/fabric/key)

    Returns:
        True if successful
    """
    status, _ = _invoke_host_fn(_FN_STATE_DELETE, path)
    return status == 0


def state_get_fabric(fabric_id):
    """Get fabric metadata.

    Args:
        fabric_id: The fabric identifier

    Returns:
        JSON string with fabric info, or None on error
    """
    status, result = _invoke_host_fn(_FN_STATE_GET_FABRIC, fabric_id)
    if status == 0:
        return result
    return None


def state_create_snapshot(path, label=''):
    """Create a snapshot of state.

    Args:
        path: State path to snapshot
        label: Optional snapshot label

    Returns:
        JSON string with snapshot metadata, or None on error
    """
    status, result = _invoke_host_fn(_FN_STATE_CREATE_SNAPSHOT, path, label)
    if status == 0:
        return result
    return None


def get_attestation(attestation_id):
    """Retrieve an attestation by ID.

    Args:
        attestation_id: The attestation ID (e.g. "att_a1b2c3...")

    Returns:
        JSON string with attestation data including proof_hash, signature, etc.,
        or None if not found
    """
    status, result = _invoke_host_fn(_FN_GET_ATTESTATION, attestation_id)
    if status == 0:
        return result
    return None


def delegate(function_id, input_data, options=None):
    """Delegate execution to another function with trust-aware routing.

    Args:
        function_id: The target function ID to delegate to
        input_data: JSON string or dict of input to pass to the target
        options: Optional JSON string or dict with delegation options:
            - min_trust_score: Minimum trust score (0-100)
            - min_trust_tier: Minimum trust tier
            - timeout_ms: Timeout in milliseconds
            - retry: Whether to retry on failure
            - max_retries: Maximum retries

    Returns:
        JSON string with execution result, or None on error
    """
    if isinstance(input_data, dict):
        input_data = json.dumps(input_data)
    if options is None:
        options = ''
    elif isinstance(options, dict):
        options = json.dumps(options)
    status, result = _invoke_host_fn(_FN_DELEGATE, function_id, input_data, options)
    if status == 0:
        return result
    return None


def is_available():
    """Check if FunctionFly host functions are available in this runtime."""
    return _is_wasm_environment()
