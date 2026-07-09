(module
  ;; IoT Memory Stress fixture - allocates memory to test limits
  (memory (export "memory") 16)

  (func (export "allocate_memory_iot") (param $kb i32) (result i32)
    ;; Attempt to grow memory by kb pages worth (64KB per page in wasm)
    (memory.grow (local.get $kb))
  )

  (func (export "get_memory_pages") (result i32)
    (memory.size)
  )
)
