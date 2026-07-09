;; IoT Error fixture
;; Exports:
;;   handler(param i32 ptr, i32 len) -> i32 (returns 11 = recoverable error)
;;   generate_error_iot(param i32 code) -> i32 (returns error code)
(module
  (memory (export "memory") 1)
  (data (i32.const 0) "{\"error\":\"timeout\",\"recoverable\":1}")

  (func (export "handler") (param $ptr i32) (param $len i32) (result i32)
    (i32.const 11)
  )

  (func (export "generate_error_iot") (param $code i32) (result i32)
    (local.get $code)
  )
)
