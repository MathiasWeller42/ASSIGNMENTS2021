define i64 @main () { 
    %result_ptr = alloca i64
    %x = icmp eq i64 1, 0
    br i1 %x, label %L1, label %L2

    ; <label>: L1
    store i64 1, i64* %result_ptr
    br label %L3
    
    ; <label>: L2
    store i64 2, i64* %result_ptr
    br label %L3
    
    ; <label>: L3
    %y = load i64*, %result_ptr
    ret i64 %y
    
}
