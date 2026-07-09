(module
  ;; IoT Actuator fixture - processes commands
  (memory (export "memory") 1)
  (data (i32.const 0) "{\"status\":\"ok\",\"action\":\"activate\"}")

  (func (export "process_command") (param $cmdPtr i32) (param $cmdLen i32) (result i32)
    ;; Write a 4-byte response code at ptr, then return length
    (i32.store (i32.const 0) (i32.const 200))
    (i32.store (i32.const 4) (i32.const 32))
    (i32.const 32)
  )

  (func (export "activate") (result i32)
    (i32.const 1)
  )

  (func (export "deactivate") (result i32)
    (i32.const 0)
  )
)
