(**************************************************************************)
(* AU compilation.                                                        *)
(* Skeleton file -- expected to be modified as part of the assignment     *)
(**************************************************************************)

open Tigercommon 
module S = Symbol 
module Ty = Types

type enventry 
  = VarEntry of Types.ty 
  | FunEntry of { formals: Types.ty list; result: Types.ty }

let printVenv = S.enter (S.empty, (S.symbol "print"), (FunEntry { formals=[STRING]; result=VOID })) (*done*)
let flushVenv = S.enter (printVenv, (S.symbol "flush"), (FunEntry { formals=[]; result=VOID })) (*done*)
let getcharVenv = S.enter (flushVenv, (S.symbol "getchar"), (FunEntry { formals=[]; result=STRING })) (*done*)
let ordVenv = S.enter (getcharVenv, (S.symbol "ord"), (FunEntry { formals=[STRING]; result=INT })) (*done*)
let chrVenv = S.enter (ordVenv, (S.symbol "chr"), (FunEntry { formals=[INT]; result=STRING })) (*done*)
let sizeVenv = S.enter (chrVenv, (S.symbol "size"), (FunEntry { formals=[STRING]; result=INT })) (*done*)
let substringVenv = S.enter (sizeVenv, (S.symbol "substring"), (FunEntry { formals=[STRING; INT; INT]; result=STRING })) (*done*) 
let concatVenv = S.enter (substringVenv, (S.symbol "concat"), (FunEntry { formals=[STRING; STRING]; result=STRING })) (*done*)
let notVenv = S.enter (concatVenv, (S.symbol "not"), (FunEntry { formals=[INT]; result=INT })) (*done*)
let baseVenv = S.enter (notVenv, (S.symbol "exit"), (FunEntry { formals=[INT]; result=VOID })) (*done*)

