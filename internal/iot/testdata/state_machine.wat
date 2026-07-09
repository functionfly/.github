(module
  ;; IoT State Machine fixture - tracks device state transitions
  (memory (export "memory") 1)
  (data (i32.const 0) "{\"state\":\"running\",\"transitions\":3}")

  (global $state (mut i32) (i32.const 0))
  (global $transitions (mut i32) (i32.const 0))

  (func (export "get_state_id") (result i32)
    (global.get $state)
  )

  (func (export "transition") (param $newState i32) (result i32)
    (global.set $state (local.get $newState))
    (global.set $transitions
      (i32.add (global.get $transitions) (i32.const 1))
    )
    (global.get $transitions)
  )
)
