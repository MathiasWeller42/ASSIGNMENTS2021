%Tuple = type { i32 , i32 }

define i32 @foo (%Tuple * %tuple ) {
%ptr = getelementptr inbounds %Tuple , %Tuple * %tuple , i64 2, i32 1
%x = load i32 , i32 * %ptr
ret i32 %x

}
