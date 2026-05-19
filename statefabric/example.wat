;; Example WebAssembly module demonstrating CommitEvent API
;; This shows how WASM functions can commit events to the StateFabric event log

(module
  ;; Import memory for shared communication
  (import "env" "memory" (memory 1))

  ;; Import host functions
  (import "env" "commit_event" (func $commit_event (param i32 i32 i32 i32 i32) (result i32)))
  (import "env" "get_state" (func $get_state (param i32 i32 i32 i32) (result i32)))
  (import "env" "set_state" (func $set_state (param i32 i32 i32 i32) (result i32)))

  ;; Global variables for string constants
  (data (i32.const 0) "user_count")
  (data (i32.const 16) "counter")
  (data (i32.const 24) "{\"count\":1}")

  ;; Function to increment a counter and commit the event
  (func $increment_counter (export "increment_counter") (result i32)
    ;; Read current counter value
    (call $get_state
      (i32.const 24) ;; "counter" key
      (i32.const 7)  ;; key length
      (i32.const 1024) ;; output buffer
      (i32.const 256)  ;; max output length
    )

    ;; For simplicity, just set a new value
    ;; In a real implementation, you'd parse the JSON and increment

    ;; Commit SET event for counter
    (call $commit_event
      (i32.const 0)   ;; SET event type
      (i32.const 24)  ;; "counter" key
      (i32.const 7)   ;; key length
      (i32.const 32)  ;; value: "{\"count\":1}"
      (i32.const 10)  ;; value length
    )

    ;; Return success
    (i32.const 0)
  )

  ;; Function to update user count
  (func $update_user_count (export "update_user_count") (param $delta i32) (result i32)
    ;; Commit MERGE event to update user count
    (call $commit_event
      (i32.const 2)   ;; MERGE event type
      (i32.const 0)   ;; "user_count" key
      (i32.const 10)  ;; key length
      (i32.const 32)  ;; value: "{\"count\":1}"
      (i32.const 10)  ;; value length
    )

    ;; Return success
    (i32.const 0)
  )

  ;; Main function that demonstrates multiple operations
  (func $process_data (export "process_data") (result i32)
    ;; Increment counter
    (call $increment_counter)
    drop

    ;; Update user count
    (call $update_user_count (i32.const 1))
    drop

    ;; Return success
    (i32.const 0)
  )
)