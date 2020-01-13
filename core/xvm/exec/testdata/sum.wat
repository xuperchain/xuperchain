(module
    (func (result i64) (local i64) (local i64) (local i64)
        i64.const 0
        set_local 0
        i64.const 100
        set_local 1
        (block
            (loop
                get_local 0
                get_local 2
                i64.add
                set_local 2

                get_local 0
                i64.const 1
                i64.add
                tee_local 0
                get_local 1
                i64.eq
                br_if 1
                br 0
            )
        )
        get_local 2
        return
    )
    (export "run" (func 0))
)
