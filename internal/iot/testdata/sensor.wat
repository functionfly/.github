(module
  ;; IoT Sensor fixture - emits a temperature/humidity JSON reading
  (memory (export "memory") 1)
  (data (i32.const 0) "{\"temp\":72.5,\"humidity\":45.0}")

  (func (export "handler") (param $ptr i32) (param $len i32) (result i32)
    ;; Return pointer to JSON string; length is at offset 0 as i32
    (i32.load (i32.const 0))
    (i32.load (i32.const 4))
    (i32.add)
  )

  (func (export "get_temperature") (result f32)
    (f32.const 72.5)
  )

  (func (export "get_humidity") (result f32)
    (f32.const 45.0)
  )
)
