;; IoT Actuator fixture
;; Exports:
;;   handler(param i32 ptr, i32 len) -> i32 (returns 200 on success)
;;   process_command(param i32 ptr, i32 len) -> i32 (returns status code)
(module
  (memory (export "memory") 1)
  (data (i32.const 0) "{\"status\":\"ok\",\"action\":\"activate\"}")

  (func (export "handler") (param $ptr i32) (param $len i32) (result i32)
    (i32.const 12)
  )

  (func (export "process_command") (param $ptr i32) (param $len i32) (result i32)
    (i32.const 200)
  )
)
