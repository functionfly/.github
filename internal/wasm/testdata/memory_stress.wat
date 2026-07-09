;; IoT Memory Stress fixture
;; Exports:
;;   handler(param i32 ptr, i32 len) -> i32
;;   allocate_memory_iot(param i32 kb) -> i32 (grow memory by kb pages)
(module
  (memory (export "memory") 1)

  (func (export "handler") (param $ptr i32) (param $len i32) (result i32)
    (i32.const 16)
  )

  (func (export "allocate_memory_iot") (param $kb i32) (result i32)
    (memory.grow (local.get $kb))
  )
)
