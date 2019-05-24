(module
  (type (;0;) (func (result i32)))
  (type (;1;) (func (param i32)))
  (type (;2;) (func (param i32) (result i32)))
  (type (;3;) (func))
  (import "env" "___setErrNo" (func (;0;) (type 1)))
  (import "env" "_print" (func (;1;) (type 1)))
  (import "env" "abortOnCannotGrowMemory" (func (;2;) (type 0)))
  (import "env" "getTotalMemory" (func (;3;) (type 0)))
  (import "env" "enlargeMemory" (func (;4;) (type 0)))
  (import "env" "DYNAMICTOP_PTR" (global (;0;) i32))
  (import "env" "STACKTOP" (global (;1;) i32))
  (import "env" "memory" (memory (;0;) 160 160))
  (func (;5;) (type 2) (param i32) (result i32)
    (local i32)
    local.get 0
    i32.const 0
    i32.gt_s
    global.get 2
    i32.load
    local.tee 1
    local.get 0
    i32.add
    local.tee 0
    local.get 1
    i32.lt_s
    i32.and
    local.get 0
    i32.const 0
    i32.lt_s
    i32.or
    if  ;; label = @1
      call 2
      drop
      i32.const 12
      call 0
      i32.const -1
      return
    end
    global.get 2
    local.get 0
    i32.store
    local.get 0
    call 3
    i32.gt_s
    if  ;; label = @1
      call 4
      i32.eqz
      if  ;; label = @2
        global.get 2
        local.get 1
        i32.store
        i32.const 12
        call 0
        i32.const -1
        return
      end
    end
    local.get 1)
  (func (;6;) (type 2) (param i32) (result i32)
    (local i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32)
    block  ;; label = @1
      block  ;; label = @2
        block  ;; label = @3
          global.get 3
          local.set 10
          global.get 3
          i32.const 16
          i32.add
          global.set 3
          local.get 10
          local.set 9
          block (result i32)  ;; label = @4
            local.get 0
            i32.const 245
            i32.lt_u
            if (result i32)  ;; label = @5
              i32.const 1024
              i32.load
              local.tee 5
              i32.const 16
              local.get 0
              i32.const 11
              i32.add
              i32.const -8
              i32.and
              local.get 0
              i32.const 11
              i32.lt_u
              select
              local.tee 2
              i32.const 3
              i32.shr_u
              local.tee 0
              i32.shr_u
              local.tee 1
              i32.const 3
              i32.and
              if  ;; label = @6
                local.get 1
                i32.const 1
                i32.and
                i32.const 1
                i32.xor
                local.get 0
                i32.add
                local.tee 0
                i32.const 3
                i32.shl
                i32.const 1064
                i32.add
                local.tee 1
                i32.const 8
                i32.add
                local.tee 4
                i32.load
                local.tee 2
                i32.const 8
                i32.add
                local.tee 6
                i32.load
                local.tee 3
                local.get 1
                i32.eq
                if  ;; label = @7
                  i32.const 1024
                  local.get 5
                  i32.const 1
                  local.get 0
                  i32.shl
                  i32.const -1
                  i32.xor
                  i32.and
                  i32.store
                else
                  local.get 3
                  local.get 1
                  i32.store offset=12
                  local.get 4
                  local.get 3
                  i32.store
                end
                local.get 2
                local.get 0
                i32.const 3
                i32.shl
                local.tee 0
                i32.const 3
                i32.or
                i32.store offset=4
                local.get 2
                local.get 0
                i32.add
                i32.const 4
                i32.add
                local.tee 0
                local.get 0
                i32.load
                i32.const 1
                i32.or
                i32.store
                local.get 10
                global.set 3
                local.get 6
                return
              end
              local.get 2
              i32.const 1032
              i32.load
              local.tee 7
              i32.gt_u
              if (result i32)  ;; label = @6
                local.get 1
                if  ;; label = @7
                  local.get 1
                  local.get 0
                  i32.shl
                  i32.const 2
                  local.get 0
                  i32.shl
                  local.tee 0
                  i32.const 0
                  local.get 0
                  i32.sub
                  i32.or
                  i32.and
                  local.tee 0
                  i32.const 0
                  local.get 0
                  i32.sub
                  i32.and
                  i32.const -1
                  i32.add
                  local.tee 1
                  i32.const 12
                  i32.shr_u
                  i32.const 16
                  i32.and
                  local.set 0
                  local.get 1
                  local.get 0
                  i32.shr_u
                  local.tee 1
                  i32.const 5
                  i32.shr_u
                  i32.const 8
                  i32.and
                  local.tee 3
                  local.get 0
                  i32.or
                  local.get 1
                  local.get 3
                  i32.shr_u
                  local.tee 0
                  i32.const 2
                  i32.shr_u
                  i32.const 4
                  i32.and
                  local.tee 1
                  i32.or
                  local.get 0
                  local.get 1
                  i32.shr_u
                  local.tee 0
                  i32.const 1
                  i32.shr_u
                  i32.const 2
                  i32.and
                  local.tee 1
                  i32.or
                  local.get 0
                  local.get 1
                  i32.shr_u
                  local.tee 0
                  i32.const 1
                  i32.shr_u
                  i32.const 1
                  i32.and
                  local.tee 1
                  i32.or
                  local.get 0
                  local.get 1
                  i32.shr_u
                  i32.add
                  local.tee 3
                  i32.const 3
                  i32.shl
                  i32.const 1064
                  i32.add
                  local.tee 0
                  i32.const 8
                  i32.add
                  local.tee 6
                  i32.load
                  local.tee 1
                  i32.const 8
                  i32.add
                  local.tee 8
                  i32.load
                  local.tee 4
                  local.get 0
                  i32.eq
                  if  ;; label = @8
                    i32.const 1024
                    local.get 5
                    i32.const 1
                    local.get 3
                    i32.shl
                    i32.const -1
                    i32.xor
                    i32.and
                    local.tee 0
                    i32.store
                  else
                    local.get 4
                    local.get 0
                    i32.store offset=12
                    local.get 6
                    local.get 4
                    i32.store
                    local.get 5
                    local.set 0
                  end
                  local.get 1
                  local.get 2
                  i32.const 3
                  i32.or
                  i32.store offset=4
                  local.get 1
                  local.get 2
                  i32.add
                  local.tee 4
                  local.get 3
                  i32.const 3
                  i32.shl
                  local.tee 3
                  local.get 2
                  i32.sub
                  local.tee 5
                  i32.const 1
                  i32.or
                  i32.store offset=4
                  local.get 1
                  local.get 3
                  i32.add
                  local.get 5
                  i32.store
                  local.get 7
                  if  ;; label = @8
                    i32.const 1044
                    i32.load
                    local.set 3
                    local.get 7
                    i32.const 3
                    i32.shr_u
                    local.tee 2
                    i32.const 3
                    i32.shl
                    i32.const 1064
                    i32.add
                    local.set 1
                    local.get 0
                    i32.const 1
                    local.get 2
                    i32.shl
                    local.tee 2
                    i32.and
                    if (result i32)  ;; label = @9
                      local.get 1
                      i32.const 8
                      i32.add
                      local.tee 2
                      i32.load
                    else
                      i32.const 1024
                      local.get 0
                      local.get 2
                      i32.or
                      i32.store
                      local.get 1
                      i32.const 8
                      i32.add
                      local.set 2
                      local.get 1
                    end
                    local.set 0
                    local.get 2
                    local.get 3
                    i32.store
                    local.get 0
                    local.get 3
                    i32.store offset=12
                    local.get 3
                    local.get 0
                    i32.store offset=8
                    local.get 3
                    local.get 1
                    i32.store offset=12
                  end
                  i32.const 1032
                  local.get 5
                  i32.store
                  i32.const 1044
                  local.get 4
                  i32.store
                  local.get 10
                  global.set 3
                  local.get 8
                  return
                end
                i32.const 1028
                i32.load
                local.tee 11
                if (result i32)  ;; label = @7
                  local.get 11
                  i32.const 0
                  local.get 11
                  i32.sub
                  i32.and
                  i32.const -1
                  i32.add
                  local.tee 1
                  i32.const 12
                  i32.shr_u
                  i32.const 16
                  i32.and
                  local.set 0
                  local.get 1
                  local.get 0
                  i32.shr_u
                  local.tee 1
                  i32.const 5
                  i32.shr_u
                  i32.const 8
                  i32.and
                  local.tee 3
                  local.get 0
                  i32.or
                  local.get 1
                  local.get 3
                  i32.shr_u
                  local.tee 0
                  i32.const 2
                  i32.shr_u
                  i32.const 4
                  i32.and
                  local.tee 1
                  i32.or
                  local.get 0
                  local.get 1
                  i32.shr_u
                  local.tee 0
                  i32.const 1
                  i32.shr_u
                  i32.const 2
                  i32.and
                  local.tee 1
                  i32.or
                  local.get 0
                  local.get 1
                  i32.shr_u
                  local.tee 0
                  i32.const 1
                  i32.shr_u
                  i32.const 1
                  i32.and
                  local.tee 1
                  i32.or
                  local.get 0
                  local.get 1
                  i32.shr_u
                  i32.add
                  i32.const 2
                  i32.shl
                  i32.const 1328
                  i32.add
                  i32.load
                  local.tee 3
                  local.set 1
                  local.get 3
                  i32.load offset=4
                  i32.const -8
                  i32.and
                  local.get 2
                  i32.sub
                  local.set 8
                  loop  ;; label = @8
                    block  ;; label = @9
                      local.get 1
                      i32.load offset=16
                      local.tee 0
                      i32.eqz
                      if  ;; label = @10
                        local.get 1
                        i32.load offset=20
                        local.tee 0
                        i32.eqz
                        br_if 1 (;@9;)
                      end
                      local.get 0
                      local.tee 1
                      local.get 3
                      local.get 1
                      i32.load offset=4
                      i32.const -8
                      i32.and
                      local.get 2
                      i32.sub
                      local.tee 0
                      local.get 8
                      i32.lt_u
                      local.tee 4
                      select
                      local.set 3
                      local.get 0
                      local.get 8
                      local.get 4
                      select
                      local.set 8
                      br 1 (;@8;)
                    end
                  end
                  local.get 3
                  local.get 2
                  i32.add
                  local.tee 12
                  local.get 3
                  i32.gt_u
                  if (result i32)  ;; label = @8
                    local.get 3
                    i32.load offset=24
                    local.set 9
                    block  ;; label = @9
                      local.get 3
                      i32.load offset=12
                      local.tee 0
                      local.get 3
                      i32.eq
                      if  ;; label = @10
                        local.get 3
                        i32.const 20
                        i32.add
                        local.tee 1
                        i32.load
                        local.tee 0
                        i32.eqz
                        if  ;; label = @11
                          local.get 3
                          i32.const 16
                          i32.add
                          local.tee 1
                          i32.load
                          local.tee 0
                          i32.eqz
                          if  ;; label = @12
                            i32.const 0
                            local.set 0
                            br 3 (;@9;)
                          end
                        end
                        loop  ;; label = @11
                          block  ;; label = @12
                            local.get 0
                            i32.const 20
                            i32.add
                            local.tee 4
                            i32.load
                            local.tee 6
                            i32.eqz
                            if  ;; label = @13
                              local.get 0
                              i32.const 16
                              i32.add
                              local.tee 4
                              i32.load
                              local.tee 6
                              i32.eqz
                              br_if 1 (;@12;)
                            end
                            local.get 4
                            local.set 1
                            local.get 6
                            local.set 0
                            br 1 (;@11;)
                          end
                        end
                        local.get 1
                        i32.const 0
                        i32.store
                      else
                        local.get 3
                        i32.load offset=8
                        local.tee 1
                        local.get 0
                        i32.store offset=12
                        local.get 0
                        local.get 1
                        i32.store offset=8
                      end
                    end
                    block  ;; label = @9
                      local.get 9
                      if  ;; label = @10
                        local.get 3
                        local.get 3
                        i32.load offset=28
                        local.tee 1
                        i32.const 2
                        i32.shl
                        i32.const 1328
                        i32.add
                        local.tee 4
                        i32.load
                        i32.eq
                        if  ;; label = @11
                          local.get 4
                          local.get 0
                          i32.store
                          local.get 0
                          i32.eqz
                          if  ;; label = @12
                            i32.const 1028
                            local.get 11
                            i32.const 1
                            local.get 1
                            i32.shl
                            i32.const -1
                            i32.xor
                            i32.and
                            i32.store
                            br 3 (;@9;)
                          end
                        else
                          local.get 9
                          i32.const 16
                          i32.add
                          local.tee 1
                          local.get 9
                          i32.const 20
                          i32.add
                          local.get 1
                          i32.load
                          local.get 3
                          i32.eq
                          select
                          local.get 0
                          i32.store
                          local.get 0
                          i32.eqz
                          br_if 2 (;@9;)
                        end
                        local.get 0
                        local.get 9
                        i32.store offset=24
                        local.get 3
                        i32.load offset=16
                        local.tee 1
                        if  ;; label = @11
                          local.get 0
                          local.get 1
                          i32.store offset=16
                          local.get 1
                          local.get 0
                          i32.store offset=24
                        end
                        local.get 3
                        i32.load offset=20
                        local.tee 1
                        if  ;; label = @11
                          local.get 0
                          local.get 1
                          i32.store offset=20
                          local.get 1
                          local.get 0
                          i32.store offset=24
                        end
                      end
                    end
                    local.get 8
                    i32.const 16
                    i32.lt_u
                    if  ;; label = @9
                      local.get 3
                      local.get 8
                      local.get 2
                      i32.add
                      local.tee 0
                      i32.const 3
                      i32.or
                      i32.store offset=4
                      local.get 3
                      local.get 0
                      i32.add
                      i32.const 4
                      i32.add
                      local.tee 0
                      local.get 0
                      i32.load
                      i32.const 1
                      i32.or
                      i32.store
                    else
                      local.get 3
                      local.get 2
                      i32.const 3
                      i32.or
                      i32.store offset=4
                      local.get 12
                      local.get 8
                      i32.const 1
                      i32.or
                      i32.store offset=4
                      local.get 12
                      local.get 8
                      i32.add
                      local.get 8
                      i32.store
                      local.get 7
                      if  ;; label = @10
                        i32.const 1044
                        i32.load
                        local.set 4
                        local.get 7
                        i32.const 3
                        i32.shr_u
                        local.tee 1
                        i32.const 3
                        i32.shl
                        i32.const 1064
                        i32.add
                        local.set 0
                        i32.const 1
                        local.get 1
                        i32.shl
                        local.tee 1
                        local.get 5
                        i32.and
                        if (result i32)  ;; label = @11
                          local.get 0
                          i32.const 8
                          i32.add
                          local.tee 2
                          i32.load
                        else
                          i32.const 1024
                          local.get 1
                          local.get 5
                          i32.or
                          i32.store
                          local.get 0
                          i32.const 8
                          i32.add
                          local.set 2
                          local.get 0
                        end
                        local.set 1
                        local.get 2
                        local.get 4
                        i32.store
                        local.get 1
                        local.get 4
                        i32.store offset=12
                        local.get 4
                        local.get 1
                        i32.store offset=8
                        local.get 4
                        local.get 0
                        i32.store offset=12
                      end
                      i32.const 1032
                      local.get 8
                      i32.store
                      i32.const 1044
                      local.get 12
                      i32.store
                    end
                    local.get 10
                    global.set 3
                    local.get 3
                    i32.const 8
                    i32.add
                    return
                  else
                    local.get 2
                  end
                else
                  local.get 2
                end
              else
                local.get 2
              end
            else
              local.get 0
              i32.const -65
              i32.gt_u
              if (result i32)  ;; label = @6
                i32.const -1
              else
                local.get 0
                i32.const 11
                i32.add
                local.tee 0
                i32.const -8
                i32.and
                local.set 1
                i32.const 1028
                i32.load
                local.tee 5
                if (result i32)  ;; label = @7
                  local.get 0
                  i32.const 8
                  i32.shr_u
                  local.tee 0
                  if (result i32)  ;; label = @8
                    local.get 1
                    i32.const 16777215
                    i32.gt_u
                    if (result i32)  ;; label = @9
                      i32.const 31
                    else
                      local.get 1
                      i32.const 14
                      local.get 0
                      local.get 0
                      i32.const 1048320
                      i32.add
                      i32.const 16
                      i32.shr_u
                      i32.const 8
                      i32.and
                      local.tee 0
                      i32.shl
                      local.tee 2
                      i32.const 520192
                      i32.add
                      i32.const 16
                      i32.shr_u
                      i32.const 4
                      i32.and
                      local.tee 3
                      local.get 0
                      i32.or
                      local.get 2
                      local.get 3
                      i32.shl
                      local.tee 0
                      i32.const 245760
                      i32.add
                      i32.const 16
                      i32.shr_u
                      i32.const 2
                      i32.and
                      local.tee 2
                      i32.or
                      i32.sub
                      local.get 0
                      local.get 2
                      i32.shl
                      i32.const 15
                      i32.shr_u
                      i32.add
                      local.tee 0
                      i32.const 7
                      i32.add
                      i32.shr_u
                      i32.const 1
                      i32.and
                      local.get 0
                      i32.const 1
                      i32.shl
                      i32.or
                    end
                  else
                    i32.const 0
                  end
                  local.set 7
                  i32.const 0
                  local.get 1
                  i32.sub
                  local.set 3
                  block  ;; label = @8
                    block  ;; label = @9
                      local.get 7
                      i32.const 2
                      i32.shl
                      i32.const 1328
                      i32.add
                      i32.load
                      local.tee 0
                      if (result i32)  ;; label = @10
                        i32.const 0
                        local.set 2
                        local.get 1
                        i32.const 0
                        i32.const 25
                        local.get 7
                        i32.const 1
                        i32.shr_u
                        i32.sub
                        local.get 7
                        i32.const 31
                        i32.eq
                        select
                        i32.shl
                        local.set 6
                        loop  ;; label = @11
                          local.get 0
                          i32.load offset=4
                          i32.const -8
                          i32.and
                          local.get 1
                          i32.sub
                          local.tee 8
                          local.get 3
                          i32.lt_u
                          if  ;; label = @12
                            local.get 8
                            if (result i32)  ;; label = @13
                              local.get 8
                              local.set 3
                              local.get 0
                            else
                              local.get 0
                              local.set 2
                              i32.const 0
                              local.set 3
                              br 4 (;@9;)
                            end
                            local.set 2
                          end
                          local.get 4
                          local.get 0
                          i32.load offset=20
                          local.tee 4
                          local.get 4
                          i32.eqz
                          local.get 4
                          local.get 0
                          i32.const 16
                          i32.add
                          local.get 6
                          i32.const 31
                          i32.shr_u
                          i32.const 2
                          i32.shl
                          i32.add
                          i32.load
                          local.tee 0
                          i32.eq
                          i32.or
                          select
                          local.set 4
                          local.get 6
                          i32.const 1
                          i32.shl
                          local.set 6
                          local.get 0
                          br_if 0 (;@11;)
                        end
                        local.get 2
                      else
                        i32.const 0
                      end
                      local.set 0
                      local.get 4
                      local.get 0
                      i32.or
                      i32.eqz
                      if  ;; label = @10
                        local.get 1
                        i32.const 2
                        local.get 7
                        i32.shl
                        local.tee 0
                        i32.const 0
                        local.get 0
                        i32.sub
                        i32.or
                        local.get 5
                        i32.and
                        local.tee 0
                        i32.eqz
                        br_if 6 (;@4;)
                        drop
                        local.get 0
                        i32.const 0
                        local.get 0
                        i32.sub
                        i32.and
                        i32.const -1
                        i32.add
                        local.tee 4
                        i32.const 12
                        i32.shr_u
                        i32.const 16
                        i32.and
                        local.set 2
                        i32.const 0
                        local.set 0
                        local.get 4
                        local.get 2
                        i32.shr_u
                        local.tee 4
                        i32.const 5
                        i32.shr_u
                        i32.const 8
                        i32.and
                        local.tee 6
                        local.get 2
                        i32.or
                        local.get 4
                        local.get 6
                        i32.shr_u
                        local.tee 2
                        i32.const 2
                        i32.shr_u
                        i32.const 4
                        i32.and
                        local.tee 4
                        i32.or
                        local.get 2
                        local.get 4
                        i32.shr_u
                        local.tee 2
                        i32.const 1
                        i32.shr_u
                        i32.const 2
                        i32.and
                        local.tee 4
                        i32.or
                        local.get 2
                        local.get 4
                        i32.shr_u
                        local.tee 2
                        i32.const 1
                        i32.shr_u
                        i32.const 1
                        i32.and
                        local.tee 4
                        i32.or
                        local.get 2
                        local.get 4
                        i32.shr_u
                        i32.add
                        i32.const 2
                        i32.shl
                        i32.const 1328
                        i32.add
                        i32.load
                        local.set 4
                      end
                      local.get 4
                      if (result i32)  ;; label = @10
                        local.get 0
                        local.set 2
                        local.get 4
                        local.set 0
                        br 1 (;@9;)
                      else
                        local.get 0
                      end
                      local.set 4
                      br 1 (;@8;)
                    end
                    local.get 2
                    local.set 4
                    local.get 3
                    local.set 2
                    loop  ;; label = @9
                      block (result i32)  ;; label = @10
                        local.get 0
                        i32.load offset=4
                        local.set 13
                        local.get 0
                        i32.load offset=16
                        local.tee 3
                        i32.eqz
                        if  ;; label = @11
                          local.get 0
                          i32.load offset=20
                          local.set 3
                        end
                        local.get 13
                      end
                      i32.const -8
                      i32.and
                      local.get 1
                      i32.sub
                      local.tee 8
                      local.get 2
                      i32.lt_u
                      local.set 6
                      local.get 8
                      local.get 2
                      local.get 6
                      select
                      local.set 2
                      local.get 0
                      local.get 4
                      local.get 6
                      select
                      local.set 4
                      local.get 3
                      if (result i32)  ;; label = @10
                        local.get 3
                        local.set 0
                        br 1 (;@9;)
                      else
                        local.get 2
                      end
                      local.set 3
                    end
                  end
                  local.get 4
                  if (result i32)  ;; label = @8
                    local.get 3
                    i32.const 1032
                    i32.load
                    local.get 1
                    i32.sub
                    i32.lt_u
                    if (result i32)  ;; label = @9
                      local.get 4
                      local.get 1
                      i32.add
                      local.tee 7
                      local.get 4
                      i32.gt_u
                      if (result i32)  ;; label = @10
                        local.get 4
                        i32.load offset=24
                        local.set 9
                        block  ;; label = @11
                          local.get 4
                          i32.load offset=12
                          local.tee 0
                          local.get 4
                          i32.eq
                          if  ;; label = @12
                            local.get 4
                            i32.const 20
                            i32.add
                            local.tee 2
                            i32.load
                            local.tee 0
                            i32.eqz
                            if  ;; label = @13
                              local.get 4
                              i32.const 16
                              i32.add
                              local.tee 2
                              i32.load
                              local.tee 0
                              i32.eqz
                              if  ;; label = @14
                                i32.const 0
                                local.set 0
                                br 3 (;@11;)
                              end
                            end
                            loop  ;; label = @13
                              block  ;; label = @14
                                local.get 0
                                i32.const 20
                                i32.add
                                local.tee 6
                                i32.load
                                local.tee 8
                                i32.eqz
                                if  ;; label = @15
                                  local.get 0
                                  i32.const 16
                                  i32.add
                                  local.tee 6
                                  i32.load
                                  local.tee 8
                                  i32.eqz
                                  br_if 1 (;@14;)
                                end
                                local.get 6
                                local.set 2
                                local.get 8
                                local.set 0
                                br 1 (;@13;)
                              end
                            end
                            local.get 2
                            i32.const 0
                            i32.store
                          else
                            local.get 4
                            i32.load offset=8
                            local.tee 2
                            local.get 0
                            i32.store offset=12
                            local.get 0
                            local.get 2
                            i32.store offset=8
                          end
                        end
                        block  ;; label = @11
                          local.get 9
                          if  ;; label = @12
                            local.get 4
                            local.get 4
                            i32.load offset=28
                            local.tee 2
                            i32.const 2
                            i32.shl
                            i32.const 1328
                            i32.add
                            local.tee 6
                            i32.load
                            i32.eq
                            if  ;; label = @13
                              local.get 6
                              local.get 0
                              i32.store
                              local.get 0
                              i32.eqz
                              if  ;; label = @14
                                i32.const 1028
                                local.get 5
                                i32.const 1
                                local.get 2
                                i32.shl
                                i32.const -1
                                i32.xor
                                i32.and
                                local.tee 0
                                i32.store
                                br 3 (;@11;)
                              end
                            else
                              local.get 9
                              i32.const 16
                              i32.add
                              local.tee 2
                              local.get 9
                              i32.const 20
                              i32.add
                              local.get 2
                              i32.load
                              local.get 4
                              i32.eq
                              select
                              local.get 0
                              i32.store
                              local.get 0
                              i32.eqz
                              if  ;; label = @14
                                local.get 5
                                local.set 0
                                br 3 (;@11;)
                              end
                            end
                            local.get 0
                            local.get 9
                            i32.store offset=24
                            local.get 4
                            i32.load offset=16
                            local.tee 2
                            if  ;; label = @13
                              local.get 0
                              local.get 2
                              i32.store offset=16
                              local.get 2
                              local.get 0
                              i32.store offset=24
                            end
                            local.get 4
                            i32.load offset=20
                            local.tee 2
                            if  ;; label = @13
                              local.get 0
                              local.get 2
                              i32.store offset=20
                              local.get 2
                              local.get 0
                              i32.store offset=24
                            end
                          end
                          local.get 5
                          local.set 0
                        end
                        block  ;; label = @11
                          local.get 3
                          i32.const 16
                          i32.lt_u
                          if  ;; label = @12
                            local.get 4
                            local.get 3
                            local.get 1
                            i32.add
                            local.tee 0
                            i32.const 3
                            i32.or
                            i32.store offset=4
                            local.get 4
                            local.get 0
                            i32.add
                            i32.const 4
                            i32.add
                            local.tee 0
                            local.get 0
                            i32.load
                            i32.const 1
                            i32.or
                            i32.store
                          else
                            local.get 4
                            local.get 1
                            i32.const 3
                            i32.or
                            i32.store offset=4
                            local.get 7
                            local.get 3
                            i32.const 1
                            i32.or
                            i32.store offset=4
                            local.get 7
                            local.get 3
                            i32.add
                            local.get 3
                            i32.store
                            local.get 3
                            i32.const 3
                            i32.shr_u
                            local.set 1
                            local.get 3
                            i32.const 256
                            i32.lt_u
                            if  ;; label = @13
                              local.get 1
                              i32.const 3
                              i32.shl
                              i32.const 1064
                              i32.add
                              local.set 0
                              i32.const 1024
                              i32.load
                              local.tee 2
                              i32.const 1
                              local.get 1
                              i32.shl
                              local.tee 1
                              i32.and
                              if (result i32)  ;; label = @14
                                local.get 0
                                i32.const 8
                                i32.add
                                local.tee 2
                                i32.load
                              else
                                i32.const 1024
                                local.get 2
                                local.get 1
                                i32.or
                                i32.store
                                local.get 0
                                i32.const 8
                                i32.add
                                local.set 2
                                local.get 0
                              end
                              local.set 1
                              local.get 2
                              local.get 7
                              i32.store
                              local.get 1
                              local.get 7
                              i32.store offset=12
                              local.get 7
                              local.get 1
                              i32.store offset=8
                              local.get 7
                              local.get 0
                              i32.store offset=12
                              br 2 (;@11;)
                            end
                            local.get 3
                            i32.const 8
                            i32.shr_u
                            local.tee 1
                            if (result i32)  ;; label = @13
                              local.get 3
                              i32.const 16777215
                              i32.gt_u
                              if (result i32)  ;; label = @14
                                i32.const 31
                              else
                                local.get 3
                                i32.const 14
                                local.get 1
                                local.get 1
                                i32.const 1048320
                                i32.add
                                i32.const 16
                                i32.shr_u
                                i32.const 8
                                i32.and
                                local.tee 1
                                i32.shl
                                local.tee 2
                                i32.const 520192
                                i32.add
                                i32.const 16
                                i32.shr_u
                                i32.const 4
                                i32.and
                                local.tee 5
                                local.get 1
                                i32.or
                                local.get 2
                                local.get 5
                                i32.shl
                                local.tee 1
                                i32.const 245760
                                i32.add
                                i32.const 16
                                i32.shr_u
                                i32.const 2
                                i32.and
                                local.tee 2
                                i32.or
                                i32.sub
                                local.get 1
                                local.get 2
                                i32.shl
                                i32.const 15
                                i32.shr_u
                                i32.add
                                local.tee 1
                                i32.const 7
                                i32.add
                                i32.shr_u
                                i32.const 1
                                i32.and
                                local.get 1
                                i32.const 1
                                i32.shl
                                i32.or
                              end
                            else
                              i32.const 0
                            end
                            local.tee 1
                            i32.const 2
                            i32.shl
                            i32.const 1328
                            i32.add
                            local.set 2
                            local.get 7
                            local.get 1
                            i32.store offset=28
                            local.get 7
                            i32.const 16
                            i32.add
                            local.tee 5
                            i32.const 0
                            i32.store offset=4
                            local.get 5
                            i32.const 0
                            i32.store
                            local.get 0
                            i32.const 1
                            local.get 1
                            i32.shl
                            local.tee 5
                            i32.and
                            i32.eqz
                            if  ;; label = @13
                              i32.const 1028
                              local.get 0
                              local.get 5
                              i32.or
                              i32.store
                              local.get 2
                              local.get 7
                              i32.store
                              local.get 7
                              local.get 2
                              i32.store offset=24
                              local.get 7
                              local.get 7
                              i32.store offset=12
                              local.get 7
                              local.get 7
                              i32.store offset=8
                              br 2 (;@11;)
                            end
                            block  ;; label = @13
                              local.get 2
                              i32.load
                              local.tee 0
                              i32.load offset=4
                              i32.const -8
                              i32.and
                              local.get 3
                              i32.eq
                              if (result i32)  ;; label = @14
                                local.get 0
                              else
                                local.get 3
                                i32.const 0
                                i32.const 25
                                local.get 1
                                i32.const 1
                                i32.shr_u
                                i32.sub
                                local.get 1
                                i32.const 31
                                i32.eq
                                select
                                i32.shl
                                local.set 2
                                loop  ;; label = @15
                                  local.get 0
                                  i32.const 16
                                  i32.add
                                  local.get 2
                                  i32.const 31
                                  i32.shr_u
                                  i32.const 2
                                  i32.shl
                                  i32.add
                                  local.tee 5
                                  i32.load
                                  local.tee 1
                                  if  ;; label = @16
                                    local.get 2
                                    i32.const 1
                                    i32.shl
                                    local.set 2
                                    local.get 1
                                    i32.load offset=4
                                    i32.const -8
                                    i32.and
                                    local.get 3
                                    i32.eq
                                    br_if 3 (;@13;)
                                    local.get 1
                                    local.set 0
                                    br 1 (;@15;)
                                  end
                                end
                                local.get 5
                                local.get 7
                                i32.store
                                local.get 7
                                local.get 0
                                i32.store offset=24
                                local.get 7
                                local.get 7
                                i32.store offset=12
                                local.get 7
                                local.get 7
                                i32.store offset=8
                                br 3 (;@11;)
                              end
                              local.set 1
                            end
                            local.get 1
                            i32.const 8
                            i32.add
                            local.tee 0
                            i32.load
                            local.tee 2
                            local.get 7
                            i32.store offset=12
                            local.get 0
                            local.get 7
                            i32.store
                            local.get 7
                            local.get 2
                            i32.store offset=8
                            local.get 7
                            local.get 1
                            i32.store offset=12
                            local.get 7
                            i32.const 0
                            i32.store offset=24
                          end
                        end
                        local.get 10
                        global.set 3
                        local.get 4
                        i32.const 8
                        i32.add
                        return
                      else
                        local.get 1
                      end
                    else
                      local.get 1
                    end
                  else
                    local.get 1
                  end
                else
                  local.get 1
                end
              end
            end
          end
          local.set 0
          i32.const 1032
          i32.load
          local.tee 2
          local.get 0
          i32.ge_u
          if  ;; label = @4
            i32.const 1044
            i32.load
            local.set 1
            local.get 2
            local.get 0
            i32.sub
            local.tee 3
            i32.const 15
            i32.gt_u
            if  ;; label = @5
              i32.const 1044
              local.get 1
              local.get 0
              i32.add
              local.tee 5
              i32.store
              i32.const 1032
              local.get 3
              i32.store
              local.get 5
              local.get 3
              i32.const 1
              i32.or
              i32.store offset=4
              local.get 1
              local.get 2
              i32.add
              local.get 3
              i32.store
              local.get 1
              local.get 0
              i32.const 3
              i32.or
              i32.store offset=4
            else
              i32.const 1032
              i32.const 0
              i32.store
              i32.const 1044
              i32.const 0
              i32.store
              local.get 1
              local.get 2
              i32.const 3
              i32.or
              i32.store offset=4
              local.get 1
              local.get 2
              i32.add
              i32.const 4
              i32.add
              local.tee 0
              local.get 0
              i32.load
              i32.const 1
              i32.or
              i32.store
            end
            br 2 (;@2;)
          end
          i32.const 1036
          i32.load
          local.tee 2
          local.get 0
          i32.gt_u
          if  ;; label = @4
            i32.const 1036
            local.get 2
            local.get 0
            i32.sub
            local.tee 2
            i32.store
            br 1 (;@3;)
          end
          i32.const 1496
          i32.load
          if (result i32)  ;; label = @4
            i32.const 1504
            i32.load
          else
            i32.const 1504
            i32.const 4096
            i32.store
            i32.const 1500
            i32.const 4096
            i32.store
            i32.const 1508
            i32.const -1
            i32.store
            i32.const 1512
            i32.const -1
            i32.store
            i32.const 1516
            i32.const 0
            i32.store
            i32.const 1468
            i32.const 0
            i32.store
            i32.const 1496
            local.get 9
            i32.const -16
            i32.and
            i32.const 1431655768
            i32.xor
            i32.store
            i32.const 4096
          end
          local.tee 1
          local.get 0
          i32.const 47
          i32.add
          local.tee 4
          i32.add
          local.tee 6
          i32.const 0
          local.get 1
          i32.sub
          local.tee 8
          i32.and
          local.tee 5
          local.get 0
          i32.le_u
          if  ;; label = @4
            br 3 (;@1;)
          end
          i32.const 1464
          i32.load
          local.tee 1
          if  ;; label = @4
            i32.const 1456
            i32.load
            local.tee 3
            local.get 5
            i32.add
            local.tee 9
            local.get 3
            i32.le_u
            local.get 9
            local.get 1
            i32.gt_u
            i32.or
            if  ;; label = @5
              br 4 (;@1;)
            end
          end
          local.get 0
          i32.const 48
          i32.add
          local.set 9
          block  ;; label = @4
            block  ;; label = @5
              i32.const 1468
              i32.load
              i32.const 4
              i32.and
              if  ;; label = @6
                i32.const 0
                local.set 2
              else
                block  ;; label = @7
                  block  ;; label = @8
                    block  ;; label = @9
                      i32.const 1048
                      i32.load
                      local.tee 1
                      i32.eqz
                      br_if 0 (;@9;)
                      i32.const 1472
                      local.set 3
                      loop  ;; label = @10
                        block  ;; label = @11
                          local.get 3
                          i32.load
                          local.tee 7
                          local.get 1
                          i32.le_u
                          if  ;; label = @12
                            local.get 7
                            local.get 3
                            i32.load offset=4
                            i32.add
                            local.get 1
                            i32.gt_u
                            br_if 1 (;@11;)
                          end
                          local.get 3
                          i32.load offset=8
                          local.tee 3
                          br_if 1 (;@10;)
                          br 2 (;@9;)
                        end
                      end
                      local.get 6
                      local.get 2
                      i32.sub
                      local.get 8
                      i32.and
                      local.tee 2
                      i32.const 2147483647
                      i32.lt_u
                      if  ;; label = @10
                        local.get 2
                        call 5
                        local.tee 1
                        local.get 3
                        i32.load
                        local.get 3
                        i32.load offset=4
                        i32.add
                        i32.eq
                        if  ;; label = @11
                          local.get 1
                          i32.const -1
                          i32.ne
                          br_if 6 (;@5;)
                        else
                          br 3 (;@8;)
                        end
                      else
                        i32.const 0
                        local.set 2
                      end
                      br 2 (;@7;)
                    end
                    i32.const 0
                    call 5
                    local.tee 1
                    i32.const -1
                    i32.eq
                    if (result i32)  ;; label = @9
                      i32.const 0
                    else
                      i32.const 1500
                      i32.load
                      local.tee 2
                      i32.const -1
                      i32.add
                      local.tee 3
                      local.get 1
                      i32.add
                      i32.const 0
                      local.get 2
                      i32.sub
                      i32.and
                      local.get 1
                      i32.sub
                      i32.const 0
                      local.get 3
                      local.get 1
                      i32.and
                      select
                      local.get 5
                      i32.add
                      local.tee 2
                      i32.const 1456
                      i32.load
                      local.tee 6
                      i32.add
                      local.set 3
                      local.get 2
                      local.get 0
                      i32.gt_u
                      local.get 2
                      i32.const 2147483647
                      i32.lt_u
                      i32.and
                      if (result i32)  ;; label = @10
                        i32.const 1464
                        i32.load
                        local.tee 8
                        if  ;; label = @11
                          local.get 3
                          local.get 6
                          i32.le_u
                          local.get 3
                          local.get 8
                          i32.gt_u
                          i32.or
                          if  ;; label = @12
                            i32.const 0
                            local.set 2
                            br 5 (;@7;)
                          end
                        end
                        local.get 2
                        call 5
                        local.tee 3
                        local.get 1
                        i32.eq
                        br_if 5 (;@5;)
                        local.get 3
                        local.set 1
                        br 2 (;@8;)
                      else
                        i32.const 0
                      end
                    end
                    local.set 2
                    br 1 (;@7;)
                  end
                  local.get 9
                  local.get 2
                  i32.gt_u
                  local.get 2
                  i32.const 2147483647
                  i32.lt_u
                  local.get 1
                  i32.const -1
                  i32.ne
                  i32.and
                  i32.and
                  i32.eqz
                  if  ;; label = @8
                    local.get 1
                    i32.const -1
                    i32.eq
                    if  ;; label = @9
                      i32.const 0
                      local.set 2
                      br 2 (;@7;)
                    else
                      br 4 (;@5;)
                    end
                    unreachable
                  end
                  local.get 4
                  local.get 2
                  i32.sub
                  i32.const 1504
                  i32.load
                  local.tee 3
                  i32.add
                  i32.const 0
                  local.get 3
                  i32.sub
                  i32.and
                  local.tee 3
                  i32.const 2147483647
                  i32.ge_u
                  br_if 2 (;@5;)
                  i32.const 0
                  local.get 2
                  i32.sub
                  local.set 4
                  local.get 3
                  call 5
                  i32.const -1
                  i32.eq
                  if (result i32)  ;; label = @8
                    local.get 4
                    call 5
                    drop
                    i32.const 0
                  else
                    local.get 3
                    local.get 2
                    i32.add
                    local.set 2
                    br 3 (;@5;)
                  end
                  local.set 2
                end
                i32.const 1468
                i32.const 1468
                i32.load
                i32.const 4
                i32.or
                i32.store
              end
              local.get 5
              i32.const 2147483647
              i32.lt_u
              if  ;; label = @6
                local.get 5
                call 5
                local.set 1
                i32.const 0
                call 5
                local.tee 3
                local.get 1
                i32.sub
                local.tee 4
                local.get 0
                i32.const 40
                i32.add
                i32.gt_u
                local.set 5
                local.get 4
                local.get 2
                local.get 5
                select
                local.set 2
                local.get 1
                i32.const -1
                i32.eq
                local.get 5
                i32.const 1
                i32.xor
                i32.or
                local.get 1
                local.get 3
                i32.lt_u
                local.get 1
                i32.const -1
                i32.ne
                local.get 3
                i32.const -1
                i32.ne
                i32.and
                i32.and
                i32.const 1
                i32.xor
                i32.or
                i32.eqz
                br_if 1 (;@5;)
              end
              br 1 (;@4;)
            end
            i32.const 1456
            i32.const 1456
            i32.load
            local.get 2
            i32.add
            local.tee 3
            i32.store
            local.get 3
            i32.const 1460
            i32.load
            i32.gt_u
            if  ;; label = @5
              i32.const 1460
              local.get 3
              i32.store
            end
            block  ;; label = @5
              i32.const 1048
              i32.load
              local.tee 5
              if  ;; label = @6
                i32.const 1472
                local.set 3
                block  ;; label = @7
                  block  ;; label = @8
                    loop  ;; label = @9
                      local.get 1
                      local.get 3
                      i32.load
                      local.tee 4
                      local.get 3
                      i32.load offset=4
                      local.tee 6
                      i32.add
                      i32.eq
                      br_if 1 (;@8;)
                      local.get 3
                      i32.load offset=8
                      local.tee 3
                      br_if 0 (;@9;)
                    end
                    br 1 (;@7;)
                  end
                  local.get 3
                  i32.const 4
                  i32.add
                  local.set 8
                  local.get 3
                  i32.load offset=12
                  i32.const 8
                  i32.and
                  i32.eqz
                  if  ;; label = @8
                    local.get 1
                    local.get 5
                    i32.gt_u
                    local.get 4
                    local.get 5
                    i32.le_u
                    i32.and
                    if  ;; label = @9
                      local.get 8
                      local.get 6
                      local.get 2
                      i32.add
                      i32.store
                      local.get 5
                      i32.const 0
                      local.get 5
                      i32.const 8
                      i32.add
                      local.tee 1
                      i32.sub
                      i32.const 7
                      i32.and
                      i32.const 0
                      local.get 1
                      i32.const 7
                      i32.and
                      select
                      local.tee 3
                      i32.add
                      local.set 1
                      i32.const 1036
                      i32.load
                      local.get 2
                      i32.add
                      local.tee 4
                      local.get 3
                      i32.sub
                      local.set 2
                      i32.const 1048
                      local.get 1
                      i32.store
                      i32.const 1036
                      local.get 2
                      i32.store
                      local.get 1
                      local.get 2
                      i32.const 1
                      i32.or
                      i32.store offset=4
                      local.get 5
                      local.get 4
                      i32.add
                      i32.const 40
                      i32.store offset=4
                      i32.const 1052
                      i32.const 1512
                      i32.load
                      i32.store
                      br 4 (;@5;)
                    end
                  end
                end
                local.get 1
                i32.const 1040
                i32.load
                i32.lt_u
                if  ;; label = @7
                  i32.const 1040
                  local.get 1
                  i32.store
                end
                local.get 1
                local.get 2
                i32.add
                local.set 4
                i32.const 1472
                local.set 3
                block  ;; label = @7
                  block  ;; label = @8
                    loop  ;; label = @9
                      local.get 3
                      i32.load
                      local.get 4
                      i32.eq
                      br_if 1 (;@8;)
                      local.get 3
                      i32.load offset=8
                      local.tee 3
                      br_if 0 (;@9;)
                    end
                    br 1 (;@7;)
                  end
                  local.get 3
                  i32.load offset=12
                  i32.const 8
                  i32.and
                  i32.eqz
                  if  ;; label = @8
                    local.get 3
                    local.get 1
                    i32.store
                    local.get 3
                    i32.const 4
                    i32.add
                    local.tee 3
                    local.get 3
                    i32.load
                    local.get 2
                    i32.add
                    i32.store
                    local.get 1
                    i32.const 0
                    local.get 1
                    i32.const 8
                    i32.add
                    local.tee 1
                    i32.sub
                    i32.const 7
                    i32.and
                    i32.const 0
                    local.get 1
                    i32.const 7
                    i32.and
                    select
                    i32.add
                    local.tee 9
                    local.get 0
                    i32.add
                    local.set 6
                    local.get 4
                    i32.const 0
                    local.get 4
                    i32.const 8
                    i32.add
                    local.tee 1
                    i32.sub
                    i32.const 7
                    i32.and
                    i32.const 0
                    local.get 1
                    i32.const 7
                    i32.and
                    select
                    i32.add
                    local.tee 2
                    local.get 9
                    i32.sub
                    local.get 0
                    i32.sub
                    local.set 3
                    local.get 9
                    local.get 0
                    i32.const 3
                    i32.or
                    i32.store offset=4
                    block  ;; label = @9
                      local.get 5
                      local.get 2
                      i32.eq
                      if  ;; label = @10
                        i32.const 1036
                        i32.const 1036
                        i32.load
                        local.get 3
                        i32.add
                        local.tee 0
                        i32.store
                        i32.const 1048
                        local.get 6
                        i32.store
                        local.get 6
                        local.get 0
                        i32.const 1
                        i32.or
                        i32.store offset=4
                      else
                        i32.const 1044
                        i32.load
                        local.get 2
                        i32.eq
                        if  ;; label = @11
                          i32.const 1032
                          i32.const 1032
                          i32.load
                          local.get 3
                          i32.add
                          local.tee 0
                          i32.store
                          i32.const 1044
                          local.get 6
                          i32.store
                          local.get 6
                          local.get 0
                          i32.const 1
                          i32.or
                          i32.store offset=4
                          local.get 6
                          local.get 0
                          i32.add
                          local.get 0
                          i32.store
                          br 2 (;@9;)
                        end
                        local.get 2
                        i32.load offset=4
                        local.tee 0
                        i32.const 3
                        i32.and
                        i32.const 1
                        i32.eq
                        if  ;; label = @11
                          local.get 0
                          i32.const -8
                          i32.and
                          local.set 7
                          local.get 0
                          i32.const 3
                          i32.shr_u
                          local.set 5
                          block  ;; label = @12
                            local.get 0
                            i32.const 256
                            i32.lt_u
                            if  ;; label = @13
                              local.get 2
                              i32.load offset=12
                              local.tee 0
                              local.get 2
                              i32.load offset=8
                              local.tee 1
                              i32.eq
                              if  ;; label = @14
                                i32.const 1024
                                i32.const 1024
                                i32.load
                                i32.const 1
                                local.get 5
                                i32.shl
                                i32.const -1
                                i32.xor
                                i32.and
                                i32.store
                              else
                                local.get 1
                                local.get 0
                                i32.store offset=12
                                local.get 0
                                local.get 1
                                i32.store offset=8
                              end
                            else
                              local.get 2
                              i32.load offset=24
                              local.set 8
                              block  ;; label = @14
                                local.get 2
                                i32.load offset=12
                                local.tee 0
                                local.get 2
                                i32.eq
                                if  ;; label = @15
                                  local.get 2
                                  i32.const 16
                                  i32.add
                                  local.tee 1
                                  i32.const 4
                                  i32.add
                                  local.tee 5
                                  i32.load
                                  local.tee 0
                                  if  ;; label = @16
                                    local.get 5
                                    local.set 1
                                  else
                                    local.get 1
                                    i32.load
                                    local.tee 0
                                    i32.eqz
                                    if  ;; label = @17
                                      i32.const 0
                                      local.set 0
                                      br 3 (;@14;)
                                    end
                                  end
                                  loop  ;; label = @16
                                    block  ;; label = @17
                                      local.get 0
                                      i32.const 20
                                      i32.add
                                      local.tee 5
                                      i32.load
                                      local.tee 4
                                      i32.eqz
                                      if  ;; label = @18
                                        local.get 0
                                        i32.const 16
                                        i32.add
                                        local.tee 5
                                        i32.load
                                        local.tee 4
                                        i32.eqz
                                        br_if 1 (;@17;)
                                      end
                                      local.get 5
                                      local.set 1
                                      local.get 4
                                      local.set 0
                                      br 1 (;@16;)
                                    end
                                  end
                                  local.get 1
                                  i32.const 0
                                  i32.store
                                else
                                  local.get 2
                                  i32.load offset=8
                                  local.tee 1
                                  local.get 0
                                  i32.store offset=12
                                  local.get 0
                                  local.get 1
                                  i32.store offset=8
                                end
                              end
                              local.get 8
                              i32.eqz
                              br_if 1 (;@12;)
                              block  ;; label = @14
                                local.get 2
                                i32.load offset=28
                                local.tee 1
                                i32.const 2
                                i32.shl
                                i32.const 1328
                                i32.add
                                local.tee 5
                                i32.load
                                local.get 2
                                i32.eq
                                if  ;; label = @15
                                  local.get 5
                                  local.get 0
                                  i32.store
                                  local.get 0
                                  br_if 1 (;@14;)
                                  i32.const 1028
                                  i32.const 1028
                                  i32.load
                                  i32.const 1
                                  local.get 1
                                  i32.shl
                                  i32.const -1
                                  i32.xor
                                  i32.and
                                  i32.store
                                  br 3 (;@12;)
                                else
                                  local.get 8
                                  i32.const 16
                                  i32.add
                                  local.tee 1
                                  local.get 8
                                  i32.const 20
                                  i32.add
                                  local.get 1
                                  i32.load
                                  local.get 2
                                  i32.eq
                                  select
                                  local.get 0
                                  i32.store
                                  local.get 0
                                  i32.eqz
                                  br_if 3 (;@12;)
                                end
                              end
                              local.get 0
                              local.get 8
                              i32.store offset=24
                              local.get 2
                              i32.const 16
                              i32.add
                              local.tee 5
                              i32.load
                              local.tee 1
                              if  ;; label = @14
                                local.get 0
                                local.get 1
                                i32.store offset=16
                                local.get 1
                                local.get 0
                                i32.store offset=24
                              end
                              local.get 5
                              i32.load offset=4
                              local.tee 1
                              i32.eqz
                              br_if 1 (;@12;)
                              local.get 0
                              local.get 1
                              i32.store offset=20
                              local.get 1
                              local.get 0
                              i32.store offset=24
                            end
                          end
                          local.get 2
                          local.get 7
                          i32.add
                          local.set 2
                          local.get 7
                          local.get 3
                          i32.add
                          local.set 3
                        end
                        local.get 2
                        i32.const 4
                        i32.add
                        local.tee 0
                        local.get 0
                        i32.load
                        i32.const -2
                        i32.and
                        i32.store
                        local.get 6
                        local.get 3
                        i32.const 1
                        i32.or
                        i32.store offset=4
                        local.get 6
                        local.get 3
                        i32.add
                        local.get 3
                        i32.store
                        local.get 3
                        i32.const 3
                        i32.shr_u
                        local.set 1
                        local.get 3
                        i32.const 256
                        i32.lt_u
                        if  ;; label = @11
                          local.get 1
                          i32.const 3
                          i32.shl
                          i32.const 1064
                          i32.add
                          local.set 0
                          i32.const 1024
                          i32.load
                          local.tee 2
                          i32.const 1
                          local.get 1
                          i32.shl
                          local.tee 1
                          i32.and
                          if (result i32)  ;; label = @12
                            local.get 0
                            i32.const 8
                            i32.add
                            local.tee 2
                            i32.load
                          else
                            i32.const 1024
                            local.get 2
                            local.get 1
                            i32.or
                            i32.store
                            local.get 0
                            i32.const 8
                            i32.add
                            local.set 2
                            local.get 0
                          end
                          local.set 1
                          local.get 2
                          local.get 6
                          i32.store
                          local.get 1
                          local.get 6
                          i32.store offset=12
                          local.get 6
                          local.get 1
                          i32.store offset=8
                          local.get 6
                          local.get 0
                          i32.store offset=12
                          br 2 (;@9;)
                        end
                        block (result i32)  ;; label = @11
                          local.get 3
                          i32.const 8
                          i32.shr_u
                          local.tee 0
                          if (result i32)  ;; label = @12
                            i32.const 31
                            local.get 3
                            i32.const 16777215
                            i32.gt_u
                            br_if 1 (;@11;)
                            drop
                            local.get 3
                            i32.const 14
                            local.get 0
                            local.get 0
                            i32.const 1048320
                            i32.add
                            i32.const 16
                            i32.shr_u
                            i32.const 8
                            i32.and
                            local.tee 0
                            i32.shl
                            local.tee 1
                            i32.const 520192
                            i32.add
                            i32.const 16
                            i32.shr_u
                            i32.const 4
                            i32.and
                            local.tee 2
                            local.get 0
                            i32.or
                            local.get 1
                            local.get 2
                            i32.shl
                            local.tee 0
                            i32.const 245760
                            i32.add
                            i32.const 16
                            i32.shr_u
                            i32.const 2
                            i32.and
                            local.tee 1
                            i32.or
                            i32.sub
                            local.get 0
                            local.get 1
                            i32.shl
                            i32.const 15
                            i32.shr_u
                            i32.add
                            local.tee 0
                            i32.const 7
                            i32.add
                            i32.shr_u
                            i32.const 1
                            i32.and
                            local.get 0
                            i32.const 1
                            i32.shl
                            i32.or
                          else
                            i32.const 0
                          end
                        end
                        local.tee 1
                        i32.const 2
                        i32.shl
                        i32.const 1328
                        i32.add
                        local.set 0
                        local.get 6
                        local.get 1
                        i32.store offset=28
                        local.get 6
                        i32.const 16
                        i32.add
                        local.tee 2
                        i32.const 0
                        i32.store offset=4
                        local.get 2
                        i32.const 0
                        i32.store
                        i32.const 1028
                        i32.load
                        local.tee 2
                        i32.const 1
                        local.get 1
                        i32.shl
                        local.tee 5
                        i32.and
                        i32.eqz
                        if  ;; label = @11
                          i32.const 1028
                          local.get 2
                          local.get 5
                          i32.or
                          i32.store
                          local.get 0
                          local.get 6
                          i32.store
                          local.get 6
                          local.get 0
                          i32.store offset=24
                          local.get 6
                          local.get 6
                          i32.store offset=12
                          local.get 6
                          local.get 6
                          i32.store offset=8
                          br 2 (;@9;)
                        end
                        block  ;; label = @11
                          local.get 0
                          i32.load
                          local.tee 0
                          i32.load offset=4
                          i32.const -8
                          i32.and
                          local.get 3
                          i32.eq
                          if (result i32)  ;; label = @12
                            local.get 0
                          else
                            local.get 3
                            i32.const 0
                            i32.const 25
                            local.get 1
                            i32.const 1
                            i32.shr_u
                            i32.sub
                            local.get 1
                            i32.const 31
                            i32.eq
                            select
                            i32.shl
                            local.set 2
                            loop  ;; label = @13
                              local.get 0
                              i32.const 16
                              i32.add
                              local.get 2
                              i32.const 31
                              i32.shr_u
                              i32.const 2
                              i32.shl
                              i32.add
                              local.tee 5
                              i32.load
                              local.tee 1
                              if  ;; label = @14
                                local.get 2
                                i32.const 1
                                i32.shl
                                local.set 2
                                local.get 1
                                i32.load offset=4
                                i32.const -8
                                i32.and
                                local.get 3
                                i32.eq
                                br_if 3 (;@11;)
                                local.get 1
                                local.set 0
                                br 1 (;@13;)
                              end
                            end
                            local.get 5
                            local.get 6
                            i32.store
                            local.get 6
                            local.get 0
                            i32.store offset=24
                            local.get 6
                            local.get 6
                            i32.store offset=12
                            local.get 6
                            local.get 6
                            i32.store offset=8
                            br 3 (;@9;)
                          end
                          local.set 1
                        end
                        local.get 1
                        i32.const 8
                        i32.add
                        local.tee 0
                        i32.load
                        local.tee 2
                        local.get 6
                        i32.store offset=12
                        local.get 0
                        local.get 6
                        i32.store
                        local.get 6
                        local.get 2
                        i32.store offset=8
                        local.get 6
                        local.get 1
                        i32.store offset=12
                        local.get 6
                        i32.const 0
                        i32.store offset=24
                      end
                    end
                    local.get 10
                    global.set 3
                    local.get 9
                    i32.const 8
                    i32.add
                    return
                  end
                end
                i32.const 1472
                local.set 3
                loop  ;; label = @7
                  block  ;; label = @8
                    local.get 3
                    i32.load
                    local.tee 4
                    local.get 5
                    i32.le_u
                    if  ;; label = @9
                      local.get 4
                      local.get 3
                      i32.load offset=4
                      i32.add
                      local.tee 6
                      local.get 5
                      i32.gt_u
                      br_if 1 (;@8;)
                    end
                    local.get 3
                    i32.load offset=8
                    local.set 3
                    br 1 (;@7;)
                  end
                end
                local.get 6
                i32.const -47
                i32.add
                local.tee 4
                i32.const 8
                i32.add
                local.set 3
                local.get 5
                local.get 4
                i32.const 0
                local.get 3
                i32.sub
                i32.const 7
                i32.and
                i32.const 0
                local.get 3
                i32.const 7
                i32.and
                select
                i32.add
                local.tee 3
                local.get 3
                local.get 5
                i32.const 16
                i32.add
                local.tee 9
                i32.lt_u
                select
                local.tee 3
                i32.const 8
                i32.add
                local.set 4
                i32.const 1048
                local.get 1
                i32.const 0
                local.get 1
                i32.const 8
                i32.add
                local.tee 8
                i32.sub
                i32.const 7
                i32.and
                i32.const 0
                local.get 8
                i32.const 7
                i32.and
                select
                local.tee 8
                i32.add
                local.tee 7
                i32.store
                i32.const 1036
                local.get 2
                i32.const -40
                i32.add
                local.tee 11
                local.get 8
                i32.sub
                local.tee 8
                i32.store
                local.get 7
                local.get 8
                i32.const 1
                i32.or
                i32.store offset=4
                local.get 1
                local.get 11
                i32.add
                i32.const 40
                i32.store offset=4
                i32.const 1052
                i32.const 1512
                i32.load
                i32.store
                local.get 3
                i32.const 4
                i32.add
                local.tee 8
                i32.const 27
                i32.store
                local.get 4
                i32.const 1472
                i64.load align=4
                i64.store align=4
                local.get 4
                i32.const 1480
                i64.load align=4
                i64.store offset=8 align=4
                i32.const 1472
                local.get 1
                i32.store
                i32.const 1476
                local.get 2
                i32.store
                i32.const 1484
                i32.const 0
                i32.store
                i32.const 1480
                local.get 4
                i32.store
                local.get 3
                i32.const 24
                i32.add
                local.set 1
                loop  ;; label = @7
                  local.get 1
                  i32.const 4
                  i32.add
                  local.tee 2
                  i32.const 7
                  i32.store
                  local.get 1
                  i32.const 8
                  i32.add
                  local.get 6
                  i32.lt_u
                  if  ;; label = @8
                    local.get 2
                    local.set 1
                    br 1 (;@7;)
                  end
                end
                local.get 3
                local.get 5
                i32.ne
                if  ;; label = @7
                  local.get 8
                  local.get 8
                  i32.load
                  i32.const -2
                  i32.and
                  i32.store
                  local.get 5
                  local.get 3
                  local.get 5
                  i32.sub
                  local.tee 4
                  i32.const 1
                  i32.or
                  i32.store offset=4
                  local.get 3
                  local.get 4
                  i32.store
                  local.get 4
                  i32.const 3
                  i32.shr_u
                  local.set 2
                  local.get 4
                  i32.const 256
                  i32.lt_u
                  if  ;; label = @8
                    local.get 2
                    i32.const 3
                    i32.shl
                    i32.const 1064
                    i32.add
                    local.set 1
                    i32.const 1024
                    i32.load
                    local.tee 3
                    i32.const 1
                    local.get 2
                    i32.shl
                    local.tee 2
                    i32.and
                    if (result i32)  ;; label = @9
                      local.get 1
                      i32.const 8
                      i32.add
                      local.tee 3
                      i32.load
                    else
                      i32.const 1024
                      local.get 3
                      local.get 2
                      i32.or
                      i32.store
                      local.get 1
                      i32.const 8
                      i32.add
                      local.set 3
                      local.get 1
                    end
                    local.set 2
                    local.get 3
                    local.get 5
                    i32.store
                    local.get 2
                    local.get 5
                    i32.store offset=12
                    local.get 5
                    local.get 2
                    i32.store offset=8
                    local.get 5
                    local.get 1
                    i32.store offset=12
                    br 3 (;@5;)
                  end
                  local.get 4
                  i32.const 8
                  i32.shr_u
                  local.tee 1
                  if (result i32)  ;; label = @8
                    local.get 4
                    i32.const 16777215
                    i32.gt_u
                    if (result i32)  ;; label = @9
                      i32.const 31
                    else
                      local.get 4
                      i32.const 14
                      local.get 1
                      local.get 1
                      i32.const 1048320
                      i32.add
                      i32.const 16
                      i32.shr_u
                      i32.const 8
                      i32.and
                      local.tee 1
                      i32.shl
                      local.tee 2
                      i32.const 520192
                      i32.add
                      i32.const 16
                      i32.shr_u
                      i32.const 4
                      i32.and
                      local.tee 3
                      local.get 1
                      i32.or
                      local.get 2
                      local.get 3
                      i32.shl
                      local.tee 1
                      i32.const 245760
                      i32.add
                      i32.const 16
                      i32.shr_u
                      i32.const 2
                      i32.and
                      local.tee 2
                      i32.or
                      i32.sub
                      local.get 1
                      local.get 2
                      i32.shl
                      i32.const 15
                      i32.shr_u
                      i32.add
                      local.tee 1
                      i32.const 7
                      i32.add
                      i32.shr_u
                      i32.const 1
                      i32.and
                      local.get 1
                      i32.const 1
                      i32.shl
                      i32.or
                    end
                  else
                    i32.const 0
                  end
                  local.tee 2
                  i32.const 2
                  i32.shl
                  i32.const 1328
                  i32.add
                  local.set 1
                  local.get 5
                  local.get 2
                  i32.store offset=28
                  local.get 5
                  i32.const 0
                  i32.store offset=20
                  local.get 9
                  i32.const 0
                  i32.store
                  i32.const 1028
                  i32.load
                  local.tee 3
                  i32.const 1
                  local.get 2
                  i32.shl
                  local.tee 6
                  i32.and
                  i32.eqz
                  if  ;; label = @8
                    i32.const 1028
                    local.get 3
                    local.get 6
                    i32.or
                    i32.store
                    local.get 1
                    local.get 5
                    i32.store
                    local.get 5
                    local.get 1
                    i32.store offset=24
                    local.get 5
                    local.get 5
                    i32.store offset=12
                    local.get 5
                    local.get 5
                    i32.store offset=8
                    br 3 (;@5;)
                  end
                  block  ;; label = @8
                    local.get 1
                    i32.load
                    local.tee 1
                    i32.load offset=4
                    i32.const -8
                    i32.and
                    local.get 4
                    i32.eq
                    if (result i32)  ;; label = @9
                      local.get 1
                    else
                      local.get 4
                      i32.const 0
                      i32.const 25
                      local.get 2
                      i32.const 1
                      i32.shr_u
                      i32.sub
                      local.get 2
                      i32.const 31
                      i32.eq
                      select
                      i32.shl
                      local.set 3
                      loop  ;; label = @10
                        local.get 1
                        i32.const 16
                        i32.add
                        local.get 3
                        i32.const 31
                        i32.shr_u
                        i32.const 2
                        i32.shl
                        i32.add
                        local.tee 6
                        i32.load
                        local.tee 2
                        if  ;; label = @11
                          local.get 3
                          i32.const 1
                          i32.shl
                          local.set 3
                          local.get 2
                          i32.load offset=4
                          i32.const -8
                          i32.and
                          local.get 4
                          i32.eq
                          br_if 3 (;@8;)
                          local.get 2
                          local.set 1
                          br 1 (;@10;)
                        end
                      end
                      local.get 6
                      local.get 5
                      i32.store
                      local.get 5
                      local.get 1
                      i32.store offset=24
                      local.get 5
                      local.get 5
                      i32.store offset=12
                      local.get 5
                      local.get 5
                      i32.store offset=8
                      br 4 (;@5;)
                    end
                    local.set 2
                  end
                  local.get 2
                  i32.const 8
                  i32.add
                  local.tee 1
                  i32.load
                  local.tee 3
                  local.get 5
                  i32.store offset=12
                  local.get 1
                  local.get 5
                  i32.store
                  local.get 5
                  local.get 3
                  i32.store offset=8
                  local.get 5
                  local.get 2
                  i32.store offset=12
                  local.get 5
                  i32.const 0
                  i32.store offset=24
                end
              else
                i32.const 1040
                i32.load
                local.tee 3
                i32.eqz
                local.get 1
                local.get 3
                i32.lt_u
                i32.or
                if  ;; label = @7
                  i32.const 1040
                  local.get 1
                  i32.store
                end
                i32.const 1472
                local.get 1
                i32.store
                i32.const 1476
                local.get 2
                i32.store
                i32.const 1484
                i32.const 0
                i32.store
                i32.const 1060
                i32.const 1496
                i32.load
                i32.store
                i32.const 1056
                i32.const -1
                i32.store
                i32.const 1076
                i32.const 1064
                i32.store
                i32.const 1072
                i32.const 1064
                i32.store
                i32.const 1084
                i32.const 1072
                i32.store
                i32.const 1080
                i32.const 1072
                i32.store
                i32.const 1092
                i32.const 1080
                i32.store
                i32.const 1088
                i32.const 1080
                i32.store
                i32.const 1100
                i32.const 1088
                i32.store
                i32.const 1096
                i32.const 1088
                i32.store
                i32.const 1108
                i32.const 1096
                i32.store
                i32.const 1104
                i32.const 1096
                i32.store
                i32.const 1116
                i32.const 1104
                i32.store
                i32.const 1112
                i32.const 1104
                i32.store
                i32.const 1124
                i32.const 1112
                i32.store
                i32.const 1120
                i32.const 1112
                i32.store
                i32.const 1132
                i32.const 1120
                i32.store
                i32.const 1128
                i32.const 1120
                i32.store
                i32.const 1140
                i32.const 1128
                i32.store
                i32.const 1136
                i32.const 1128
                i32.store
                i32.const 1148
                i32.const 1136
                i32.store
                i32.const 1144
                i32.const 1136
                i32.store
                i32.const 1156
                i32.const 1144
                i32.store
                i32.const 1152
                i32.const 1144
                i32.store
                i32.const 1164
                i32.const 1152
                i32.store
                i32.const 1160
                i32.const 1152
                i32.store
                i32.const 1172
                i32.const 1160
                i32.store
                i32.const 1168
                i32.const 1160
                i32.store
                i32.const 1180
                i32.const 1168
                i32.store
                i32.const 1176
                i32.const 1168
                i32.store
                i32.const 1188
                i32.const 1176
                i32.store
                i32.const 1184
                i32.const 1176
                i32.store
                i32.const 1196
                i32.const 1184
                i32.store
                i32.const 1192
                i32.const 1184
                i32.store
                i32.const 1204
                i32.const 1192
                i32.store
                i32.const 1200
                i32.const 1192
                i32.store
                i32.const 1212
                i32.const 1200
                i32.store
                i32.const 1208
                i32.const 1200
                i32.store
                i32.const 1220
                i32.const 1208
                i32.store
                i32.const 1216
                i32.const 1208
                i32.store
                i32.const 1228
                i32.const 1216
                i32.store
                i32.const 1224
                i32.const 1216
                i32.store
                i32.const 1236
                i32.const 1224
                i32.store
                i32.const 1232
                i32.const 1224
                i32.store
                i32.const 1244
                i32.const 1232
                i32.store
                i32.const 1240
                i32.const 1232
                i32.store
                i32.const 1252
                i32.const 1240
                i32.store
                i32.const 1248
                i32.const 1240
                i32.store
                i32.const 1260
                i32.const 1248
                i32.store
                i32.const 1256
                i32.const 1248
                i32.store
                i32.const 1268
                i32.const 1256
                i32.store
                i32.const 1264
                i32.const 1256
                i32.store
                i32.const 1276
                i32.const 1264
                i32.store
                i32.const 1272
                i32.const 1264
                i32.store
                i32.const 1284
                i32.const 1272
                i32.store
                i32.const 1280
                i32.const 1272
                i32.store
                i32.const 1292
                i32.const 1280
                i32.store
                i32.const 1288
                i32.const 1280
                i32.store
                i32.const 1300
                i32.const 1288
                i32.store
                i32.const 1296
                i32.const 1288
                i32.store
                i32.const 1308
                i32.const 1296
                i32.store
                i32.const 1304
                i32.const 1296
                i32.store
                i32.const 1316
                i32.const 1304
                i32.store
                i32.const 1312
                i32.const 1304
                i32.store
                i32.const 1324
                i32.const 1312
                i32.store
                i32.const 1320
                i32.const 1312
                i32.store
                i32.const 1048
                local.get 1
                i32.const 0
                local.get 1
                i32.const 8
                i32.add
                local.tee 3
                i32.sub
                i32.const 7
                i32.and
                i32.const 0
                local.get 3
                i32.const 7
                i32.and
                select
                local.tee 3
                i32.add
                local.tee 5
                i32.store
                i32.const 1036
                local.get 2
                i32.const -40
                i32.add
                local.tee 2
                local.get 3
                i32.sub
                local.tee 3
                i32.store
                local.get 5
                local.get 3
                i32.const 1
                i32.or
                i32.store offset=4
                local.get 1
                local.get 2
                i32.add
                i32.const 40
                i32.store offset=4
                i32.const 1052
                i32.const 1512
                i32.load
                i32.store
              end
            end
            i32.const 1036
            i32.load
            local.tee 1
            local.get 0
            i32.gt_u
            if  ;; label = @5
              i32.const 1036
              local.get 1
              local.get 0
              i32.sub
              local.tee 2
              i32.store
              br 2 (;@3;)
            end
          end
          i32.const 1520
          i32.const 12
          i32.store
          br 2 (;@1;)
        end
        i32.const 1048
        i32.const 1048
        i32.load
        local.tee 1
        local.get 0
        i32.add
        local.tee 3
        i32.store
        local.get 3
        local.get 2
        i32.const 1
        i32.or
        i32.store offset=4
        local.get 1
        local.get 0
        i32.const 3
        i32.or
        i32.store offset=4
      end
      local.get 10
      global.set 3
      local.get 1
      i32.const 8
      i32.add
      return
    end
    local.get 10
    global.set 3
    i32.const 0)
  (func (;7;) (type 3)
    i32.const 1024
    call 6
    call 1)
  (func (;8;) (type 0) (result i32)
    i32.const 1520)
  (global (;2;) (mut i32) (global.get 0))
  (global (;3;) (mut i32) (global.get 1))
  (export "___errno_location" (func 8))
  (export "_run" (func 7)))
