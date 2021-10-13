@x = global i64 42
@s = global [5 x i8] c"Hello"

define i64 @main() {
    %y = load i64, i64* @x
    %f = call i64 @foo (i64 %y)
    ret %f
}

define i64 @foo (i64 %x) {
    %y = mul i64 %x, 3
    ret i64 %y
}