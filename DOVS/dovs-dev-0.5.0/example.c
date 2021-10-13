struct Tuple {
int x;
int y;
};

int foo ( struct Tuple *tuple ) {
return tuple [2]. y;
// Try : return tuple [0]. y
// return tuple ->y
}

int main ( int argc , char **argv ) {
struct Tuple tuples [] = { {11 , 22} ,
{33 , 44} ,
{55 , 66} };
return foo ( tuples ); // returns 66
}