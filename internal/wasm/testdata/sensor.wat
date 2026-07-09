;; IoT Sensor fixture
;; Exports:
;;   handler(param i32 ptr, i32 len) -> i32 (returns JSON output length)
;;   get_temperature() -> f32 (returns 72.5)
(module
  (memory (export "memory") 1)
  (data (i32.const 0) "{\"temp\":72.5,\"humidity\":45.0}")

  (func (export "handler") (param $ptr i32) (param $len i32) (result i32)
    (i32.const 27)
  )

  (func (export "get_temperature") (result f32)
    (f32.const 72.5)
  )
)
