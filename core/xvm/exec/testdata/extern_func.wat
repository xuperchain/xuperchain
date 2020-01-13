(module
  (type (;0;) (func (param i32)))
  (type (;1;) (func))
  (import "env" "_print" (func (;0;) (type 0)))
  (import "env" "memory" (memory (;0;) 160 160))
  (func (;1;) (type 1)
    i32.const 1024
    call 0)
  (export "_run" (func 1))
  (data (;0;) (i32.const 1024) "hello world"))
