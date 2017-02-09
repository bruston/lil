:main
    var max
    push_int64 10
    store max
    push_int64 0
    :loop
        dup
        print
        inc
        dup
        load max
        jump_lt loop
    halt
