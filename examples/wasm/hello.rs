// FunctionFly example: Hello from Rust
// Build: cargo build --release --target wasm32-wasip1
// Manifest: {"name":"hello-rust","version":"1.0.0","runtime":"rust","entry":"hello.rs"}

#[no_mangle]
pub extern "C" fn init() {}

#[no_mangle]
pub extern "C" fn execute(ptr: *const u8, len: usize) -> *const u8 {
    let input = unsafe {
        let slice = std::slice::from_raw_parts(ptr, len);
        std::str::from_utf8_unchecked(slice)
    };

    let response = format!(
        r#"{{"message": "Hello from Rust!", "input_length": {}, "ok": true}}"#,
        input.len()
    );

    let boxed = response.into_bytes().leak();
    boxed.as_ptr()
}

#[no_mangle]
pub extern "C" fn alloc(size: usize) -> *mut u8 {
    let mut buf = Vec::with_capacity(size);
    let ptr = buf.as_mut_ptr();
    std::mem::forget(buf);
    ptr
}

#[no_mangle]
pub extern "C" fn dealloc(ptr: *mut u8, size: usize) {
    unsafe {
        let _ = Vec::from_raw_parts(ptr, size, size);
    }
}
