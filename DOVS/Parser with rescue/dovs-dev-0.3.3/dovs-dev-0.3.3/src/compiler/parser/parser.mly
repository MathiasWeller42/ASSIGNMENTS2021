(**************************************************************************)
(* AU compilation.                                                        *)
(* Skeleton file -- expected to be modified as part of the assignment     *)
(* Do not distribute                                                      *)
(**************************************************************************)

%{
  open Tigercommon.Absyn   
  open ParserAux 
  open Tigercommon.Symbol

  exception ArrayError of string

%}

%token EOF
%token <string> ID
%token <int> INT 
%token <string> STRING 
%token COMMA COLON SEMICOLON 
%token LPAREN RPAREN LBRACK RBRACK LBRACE RBRACE 
%token DOT PLUS MINUS TIMES DIVIDE EQ NEQ LT LE GT GE 
%token AND OR ASSIGN ARRAY IF THEN ELSE WHILE FOR TO DO
%token LET IN END OF BREAK NIL FUNCTION VAR TYPE CARET 

%nonassoc ASSIGN DO THEN OF (*<-- Noodle-ass weak-looking motherfuckers (aka low precedence)*)
%right ELSE
%left AND OR
%nonassoc EQ NEQ LT LE GT GE 
%right FUNCTION TYPE 
%left PLUS MINUS
%left TIMES DIVIDE
%right CARET (*<-- biggest coolest dude in town (aka high precedence) *)

%start <Tigercommon.Absyn.exp> program  
(* Observe that we need to use fully qualified types for the start symbol *)

%%
(* Expressions *)
exp_base:
| NIL                                                   { NilExp }
| i=INT                                                 { IntExp i }
| s=STRING                                              { StringExp s }
| MINUS exp1=exp                                        { OpExp { left=((IntExp 0) ^! $startpos); oper=MinusOp; right=exp1} } %prec CARET
| lvalue1=lvalue                                        { VarExp (lvalue1) }
| exp1=exp op=binop exp2=exp                            { OpExp { left=exp1; oper=op; right=exp2 } } 
| IF exp1=exp THEN exp2=exp exp3=if_exp                 { IfExp { test=exp1; thn=exp2; els=exp3 } }
| LPAREN exp=seq_exp RPAREN                             { SeqExp exp }
| LET declList=decl_exp IN exp1=seq_exp END             { LetExp { decls=declList; body= ((SeqExp exp1) ^! $startpos(exp1)) } }
| lvalue1=lvalue ASSIGN exp1=exp                        { AssignExp { var=lvalue1; exp=exp1 } }
| WHILE cond=exp DO body=exp                            { WhileExp { test=cond; body=body } }
| FOR i=ID ASSIGN exp1=exp TO exp2=exp DO exp3=exp      { ForExp { var=(symbol i); escape=ref true; lo=exp1; hi=exp2; body=exp3 } }
| BREAK                                                 { BreakExp }  
| id=ID LPAREN explist=exp_list RPAREN                  { CallExp { func=(symbol id); args=explist } }     
| exp1=exp AND exp2=exp                                 { IfExp { test=exp1; thn=exp2; els=Some((IntExp 0 ^! $startpos)) } } 
| exp1=exp OR exp2=exp                                  { IfExp { test=exp1; thn=(IntExp 1 ^! $startpos); els=(Some exp2) } }
| id=ID LBRACE data=recorddata RBRACE                   { RecordExp { fields=data; typ=(symbol id) } }
| lvalue1=lvalue OF exp2=exp                            { match lvalue1 with 
                                                          | Var { var_base; _ } -> (match var_base with
                                                                                | SubscriptVar (var1, exp1) -> (match var1 with 
                                                                                                                | Var { var_base; _ } -> (match var_base with
                                                                                                                                           | SimpleVar s -> ArrayExp { typ=s; size=exp1; init=exp2 }
                                                                                                                                           | _ -> raise (ArrayError "Error when creating array, expected ID and [n]")))       
                                                                                | _ -> raise (ArrayError "Error when creating array, expected ID and [n]"))                                                                                                                                                               
                                                        }



lvalue: 
| id=ID lvalue1=lvalue_list                        { (makeLvaluePartSpec (SimpleVar (symbol id) ^@ $startpos) $startpos lvalue1) }                                

lvalue_list:
| lvalue1=lvalue_list DOT id=ID                    { List.append lvalue1 [FieldPart(symbol id)]  }
| lvalue1=lvalue_list LBRACK exp1=exp RBRACK       { List.append lvalue1 [SubscriptPart exp1] }
|                                                  { [] }

recorddata: 
|                                                  { [] }
| id=ID EQ exp1=exp                                { ((symbol id), exp1) :: [] }
| id=ID EQ exp1=exp COMMA recorddata1=recorddata   { ((symbol id), exp1) :: recorddata1 }

%inline if_exp:
| ELSE exp3=exp                                    { Some exp3 }
|                                                  { None }

decl:
| FUNCTION  fundecldata1=fundecldata               { FunctionDec (fundecldata1 ) }
| VAR i=ID ASSIGN exp1=exp                         { VarDec { name=symbol i; escape=ref true; typ=None; init=exp1; pos= $startpos } }
| VAR i=ID COLON j=ID ASSIGN exp1=exp              { VarDec { name=symbol i; escape=ref true; typ=Some(symbol j, $startpos(j)); init=exp1; pos= $startpos } }
| TYPE tydecldata1=tydecldata                      { TypeDec ( tydecldata1 ) }

tydecldata:
| i=ID EQ ty1=ty TYPE tydecldata1=tydecldata  { Tdecl { name=(symbol i); ty=ty1; pos= $startpos } :: tydecldata1 }
| i=ID EQ ty1=ty                              { Tdecl { name=(symbol i); ty=ty1; pos= $startpos } :: [] }

ty:
| j=ID                                        { NameTy (symbol j, $startpos) }
| LBRACE paramlist=fielddata RBRACE           { RecordTy paramlist }
| ARRAY OF j=ID                               { ArrayTy (symbol j, $startpos) }

fundecldata:
| id=ID LPAREN paramlist=fielddata RPAREN EQ body=exp FUNCTION fundecldata1=fundecldata             { Fdecl { name= symbol id; params=paramlist ; result=None; body=body ; pos= $startpos } :: fundecldata1 }
| id=ID LPAREN paramlist=fielddata RPAREN COLON j=ID EQ body=exp FUNCTION fundecldata1=fundecldata  { Fdecl { name= symbol id; params=paramlist ; result=Some(symbol j, $startpos(j)); body=body ; pos= $startpos } :: fundecldata1 }
| id=ID LPAREN paramlist=fielddata RPAREN EQ body=exp                                               { Fdecl { name= symbol id; params=paramlist ; result=None; body=body ; pos= $startpos } :: [] }
| id=ID LPAREN paramlist=fielddata RPAREN COLON j=ID EQ body=exp                                    { Fdecl { name= symbol id; params=paramlist ; result=Some(symbol j, $startpos(j)); body=body ; pos= $startpos } :: [] }


fielddata:
| id=ID COLON j=ID COMMA fielddata1=fielddata     { Field { name=(symbol id); escape=ref true; typ=(symbol j, $startpos(j)); pos= $startpos } :: fielddata1 }
| id=ID COLON j=ID                                { Field { name=(symbol id); escape=ref true; typ=(symbol j, $startpos(j)); pos= $startpos } :: [] }
|                                                 { [] }

decl_exp:
| decl1=decl decls=decl_exp           { decl1 :: decls }
|                                     { [] }

exp_list:
|                                     { [] }
| exp1=exp                            { exp1 :: [] }
| exp1=exp COMMA exp2=exp_list        { exp1 :: exp2 }

seq_exp: 
|                                     { [] }
| exp1=exp                            { exp1 :: [] }
| exp1=exp SEMICOLON exp2=seq_exp     { exp1 :: exp2 }

%inline binop:
| PLUS    { PlusOp }
| MINUS   { MinusOp }
| TIMES   { TimesOp }
| DIVIDE  { DivideOp }
| EQ      { EqOp }
| NEQ     { NeqOp }
| LT      { LtOp }
| LE      { LeOp }
| GT      { GtOp }
| GE      { GeOp }
| CARET   { ExponentOp }


(* Top-level *)
program: e = exp EOF { e }

exp:
| e=exp_base  { e ^! $startpos }



