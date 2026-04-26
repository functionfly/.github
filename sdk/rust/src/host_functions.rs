//! Host function bindings — these are provided by the FunctionFly WASM runtime.

extern "C" {
    fn functionfly_log(msg: *const u8, len: i32);
    fn functionfly_get_env(key: *const u8, key_len: i32, buf: *mut u8, buf_len: i32) -> i32;
    fn functionfly_fetch(url: *const u8, url_len: i32, buf: *mut u8, buf_len: i32) -> i32;
    fn functionfly_kv_get(key: *const u8, key_len: i32, buf: *mut u8, buf_len: i32) -> i32;
    fn functionfly_kv_set(key: *const u8, key_len: i32, val: *const u8, val_len: i32) -> i32;
    fn functionfly_ai(prompt: *const u8, prompt_len: i32, buf: *mut u8, buf_len: i32) -> i32;
}

/// Log a message to the FunctionFly logging system.
pub fn log(msg: &str) {
    unsafe {
        functionfly_log(msg.as_ptr(), msg.len() as i32);
    }
}

/// Get an environment variable from the runtime.
pub fn get_env(key: &str) -> Option<String> {
    let mut buf = vec![0u8; 1024];
    let len = unsafe {
        functionfly_get_env(key.as_ptr(), key.len() as i32, buf.as_mut_ptr(), buf.len() as i32)
    };
    if len > 0 {
        buf.truncate(len as usize);
        String::from_utf8(buf).ok()
    } else {
        None
    }
}

/// Fetch a URL via the FunctionFly HTTP proxy.
pub fn fetch(url: &str) -> Result<String, String> {
    let mut buf = vec![0u8; 65536];
    let len = unsafe {
        functionfly_fetch(url.as_ptr(), url.len() as i32, buf.as_mut_ptr(), buf.len() as i32)
    };
    if len > 0 {
        buf.truncate(len as usize);
        String::from_utf8(buf).map_err(|e| e.to_string())
    } else {
        Err("fetch failed".to_string())
    }
}

/// Get a value from the key-value store.
pub fn kv_get(key: &str) -> Option<String> {
    let mut buf = vec![0u8; 4096];
    let len = unsafe {
        functionfly_kv_get(key.as_ptr(), key.len() as i32, buf.as_mut_ptr(), buf.len() as i32)
    };
    if len > 0 {
        buf.truncate(len as usize);
        String::from_utf8(buf).ok()
    } else {
        None
    }
}

/// Set a value in the key-value store.
pub fn kv_set(key: &str, value: &str) -> bool {
    let len = unsafe {
        functionfly_kv_set(key.as_ptr(), key.len() as i32, value.as_ptr(), value.len() as i32)
    };
    len >= 0
}
