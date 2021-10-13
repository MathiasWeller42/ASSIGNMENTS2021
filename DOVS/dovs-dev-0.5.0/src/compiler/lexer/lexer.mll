(**************************************************************************)
(* AU compilation.                                                        *)
(* Skeleton file -- expected to be modified as part of the assignment     *)
(* Do not distribute                                                      *)
(**************************************************************************)

{
  open Tigerparser.Parser  
  exception Error of string
  let error lexbuf msg =
    let position = Lexing.lexeme_start_p lexbuf in
    let err_str = Printf.sprintf "Lexing error in file %s at position %d:%d\n"
                  position.pos_fname position.pos_lnum (position.pos_cnum - position.pos_bol + 1)
                  ^ msg ^ "\n" in
    raise (Error err_str)
}

(* Named regexps *)
let digits = ['0'-'9']+
let identifier = (['a'-'z']|['A'-'Z'])(['a'-'z']|['A'-'Z']|['0'-'9']|'_')*
let letters = (['a'-'z']|['A'-'Z'])+
let notid = ((['0'-'9']|'_')+)((['a'-'z']|['A'-'Z'])+)

(* Entrypoint *)
rule token = parse
| [' ' '\t' ]     { token lexbuf }     (* skip blanks *)
| eof                 { EOF }
| ','                 { COMMA }
| ':'                 { COLON }
| ';'                 { SEMICOLON }
| '('                 { LPAREN }
| ')'                 { RPAREN }
| '['                 { LBRACK }
| ']'                 { RBRACK }
| '{'                 { LBRACE }
| '}'                 { RBRACE }
| '.'                 { DOT }
| '+'                 { PLUS }
| '-'                 { MINUS }
| '*'                 { TIMES }
| '/'                 { DIVIDE }
| '^'                 { CARET }
| '='                 { EQ }
| "<>"                { NEQ }
| '<'                 { LT }
| "<="                { LE }
| '>'                 { GT }
| ">="                { GE }
| '&'                 { AND }
| '|'                 { OR }
| ":="                { ASSIGN }
| "while"             { WHILE }
| "for"               { FOR }
| "to"                { TO }
| "break"             { BREAK }
| "let"               { LET }
| "in"                { IN }
| "end"               { END }
| "function"          { FUNCTION }
| "var"               { VAR }
| "type"              { TYPE }
| "array"             { ARRAY }
| "if"                { IF }
| "then"              { THEN }
| "else"              { ELSE }
| "do"                { DO }
| "of"                { OF }
| "nil"               { NIL }
| digits as i         { let intopt = int_of_string_opt i in match intopt with None -> error lexbuf "Too big of an int" | Some i -> INT(i) } 
| identifier as id    { ID (id) }
| notid               { error lexbuf "An ID cannot start with a digit or underscore" }
| "/*"                { nestedComment 0 lexbuf } (*Nested comments*)
| "\n"                { Lexing.new_line lexbuf; token lexbuf }
| '"'                 { let pos = lexbuf.Lexing.lex_start_p in let buffer = Buffer.create 1 in strings buffer pos lexbuf }  (*Strings*)
| _ as t              { error lexbuf ("Invalid character '" ^ (String.make 1 t) ^ "'") }

and nestedComment commentLevel = parse (*Nested comments*)
| "/*"                { nestedComment (commentLevel+1 ) lexbuf }
| "*/"                { (if commentLevel = 0 then token else nestedComment (commentLevel-1) ) lexbuf }
| "\n"                { Lexing.new_line lexbuf; nestedComment commentLevel lexbuf }
| eof                 { error lexbuf ("Met end of file in open comment") }
| _                   { nestedComment commentLevel lexbuf }

and strings buffer pos = parse  (*Strings*)
| '"'                 { lexbuf.Lexing.lex_start_p <- pos; STRING (Buffer.contents buffer) }
| "\\n"               { Buffer.add_char buffer '\n'; strings buffer pos lexbuf } 
| "\\t"               { Buffer.add_char buffer '\t'; strings buffer pos lexbuf } 
| "\\" '"'            { Buffer.add_char buffer '"'; strings buffer pos lexbuf }     
| "\\" "\\"           { Buffer.add_char buffer '\\'; strings buffer pos lexbuf }  
| "\\" ['0'-'9']['0'-'9']['0'-'9'] as code {let c = String.sub code 1 3 (*ASCII codes*)
                                        in let ascii = int_of_string c in 
                                        if ascii <= 255 
                                        then (Buffer.add_char buffer (Char.chr ascii); strings buffer pos lexbuf) 
                                        else error lexbuf ("Illegal control character")  }
| "\\" "^" (['A'-'Z']|'['|'\\'|']'|'^'|'_'|'@') as control (*Capitalized control characters and symbols*)
                                        {let c = String.get control 2 
                                        in let ch = Char.code c 
                                        in Buffer.add_char buffer (Char.chr(ch - 64)); 
                                        strings buffer pos lexbuf  } 
| "\\" "^" ['a'-'z'] as control         {let c = String.get control 2 (*Lower case control characters*)
                                        in let ch = Char.code c 
                                        in  Buffer.add_char buffer (Char.chr(ch - 96)); 
                                        strings buffer pos lexbuf  } 
| "\\" "^" '?'        {Buffer.add_char buffer (Char.chr(127)); strings buffer pos lexbuf  } (*'?' control character*)
| "\\" "^" _          { error lexbuf ("Illegal control character") }
| "\\"                { multiString buffer pos lexbuf } (*Start of multiline string*)
| "\n"                { error lexbuf ("Multiline string without multiline marker") }
| eof                 { error lexbuf ("Met end of file in open string") }
| _ as s              {let ascii = Char.code s in if ascii <= 127 && ascii >=32 then (Buffer.add_char buffer (Char.chr(ascii)); strings buffer pos lexbuf) else error lexbuf ("Illegal character") }

and multiString buffer pos = parse (*Multiline string*)
| "\n"                { Lexing.new_line lexbuf; multiString buffer pos lexbuf }
| [' ' '\t']          { multiString buffer pos lexbuf }
| "\\"                { strings buffer pos lexbuf }
| eof                 { error lexbuf ("Met end of file in open string") }
| _ as s              { error lexbuf ("Illegal control character" ^ (String.make 1 s) ^ "'") }
