;; IoT State Machine fixture
;; Exports:
;;   handler(param i32 ptr, i32 len) -> i32
;;   get_state_id() -> i32 (returns current state)
(module
  (memory (export "memory") 1)
  (data (i32.const 0) "{\"state\":\"running\",\"transitions\":3}")

  (global $state (mut i32) (i32.const 1))

  (func (export "handler") (param $ptr i32) (param $len i32) (result i32)
    (i32.const 7)
  )

  (func (export "get_state_id") (result i32)
    (global.get $state)
  )
)
