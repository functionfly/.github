(module
  ;; IoT Error fixture - simulates error handling
  (memory (export "memory") 1)
  (data (i32.const 0) "{\"error\":\"timeout\",\"recoverable\":1}")

  (func (export "generate_error_iot") (param $code i32) (result i32)
    (local.get $code)
  )

  (func (export "is_recoverable") (result i32)
    (i32.const 1)
  )
)
