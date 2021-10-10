(**************************************************************************)
(* AU compilation.                                                        *)
(* Skeleton file -- expected to be modified as part of the assignment     *)
(**************************************************************************)
open Tigercommon
module S = Symbol
module A = Absyn
module E = Semenv
module Err = Errenv
module EFmt = ErrorFormatter
module Ty = Types
module PT = Prtypes
module TA = Tabsyn

(** Context record contains the environments we use in our translation *)

type context =
  { venv: E.enventry S.table (* Γ from our formal typing rules *)
  ; err: Err.errenv (* error environment *) }

(*exception NotImplemented*)
(* the final code should work without this exception *)

exception NotSem0 (* for handling AST cases outside of Sem 0 feature set *)

(*Tiger Cub - (BJ)ÆFF YEAH!*)

open Ty
(*Our getType implementation*)
let getType (e: TA.exp) : Types.ty =
  match e with 
  | TA.Exp { exp_base=_; pos=_; ty } -> ty

(*transExp is the type-checker function *)
let rec transExp ({err; venv} : context) e =
  let rec trexp (A.Exp {exp_base; pos}) =
    let (^!) exp_base ty = TA.Exp {exp_base;pos;ty} in
    match exp_base with
    | A.IntExp n ->  TA.IntExp n ^! INT 
    | A.StringExp s -> TA.StringExp s ^! STRING 
    (* the above cases have been implemented in class *)
    | A.OpExp {left; oper; right} -> 
        let transLeft = (trexp left) in
        let typeLeft = getType transLeft in
        let transRight = (trexp right) in
        let typeRight = getType transRight in
          (match oper with
            | Oper.PlusOp | Oper.MinusOp | Oper.TimesOp | Oper.DivideOp | Oper.ExponentOp  -> if (typeLeft == INT && typeRight == INT) then( TA.OpExp {left=transLeft; oper=oper; right=transRight} ^! getType(transLeft)) else (Err.error err pos EFmt.errorArith; TA.ErrorExp ^! ERROR)
            | Oper.EqOp | Oper.NeqOp | Oper.LtOp | Oper.LeOp | Oper.GtOp | Oper.GeOp  -> if ((typeLeft == INT && typeRight == INT) || (typeLeft == STRING && typeRight == STRING)) then TA.OpExp {left=transLeft; oper=oper; right=transRight} ^! INT else (Err.error err pos (EFmt.errorEqNeqComparison typeLeft typeRight); TA.ErrorExp ^! ERROR)
          )       
    | A.CallExp {func; args} -> 
      (let funcLookup = S.look (venv, func) in
        (match funcLookup with
         | Some(FunEntry {formals; result}) -> (
            (if List.length args = List.length formals then 
              let new_args = ref [] in
              for i = 0 to (List.length args)-1 do 
                let transArg = trexp (List.nth args i) in
                let typeArg = getType transArg in
                (if (List.nth formals i) != typeArg 
                  then (Err.error err pos "Parameter type does not match argument type";  new_args := !new_args @ [TA.ErrorExp ^! ERROR])
                else 
                  new_args := !new_args @ [transArg])
              done;
              TA.CallExp {func; args= !new_args} ^! result 
            else (Err.error err pos (EFmt.errorNumberFunctionArguments func formals args) ; TA.ErrorExp ^! ERROR)))
         | Some(VarEntry _) -> Err.error err pos (EFmt.errorUsingVariableAsFunction func); TA.ErrorExp ^! ERROR
         | None -> Err.error err pos (EFmt.errorFunctionUndefined func); TA.ErrorExp ^! ERROR)) 

    | A.SeqExp exps -> (
      let new_exps = ref [] in
      let last_type = ref VOID in (
      for i = 0 to (List.length exps)-1 do 
        let transExp = trexp (List.nth exps i) in
        new_exps := !new_exps @ [transExp];
        if i = (List.length exps)-1 then last_type := (getType transExp) 
      done;
      if List.length !new_exps = 1 then (List.nth !new_exps 0) else (TA.SeqExp !new_exps ^! !last_type)
      ))

    | A.IfExp {test; thn; els=Some el} ->
        let transTest = (trexp test) in
        let typeTest = getType transTest in
        let transThen = (trexp thn) in
        let typeThen = getType transThen in
        let transElse = (trexp el) in
        let typeElse = getType transElse in 
          (
            if typeTest = INT then (
              if typeThen = typeElse then (
                TA.IfExp { test=transTest; thn=transThen; els=(Some transElse) } ^! typeThen
              ) else (Err.error err pos (EFmt.errorIfBranchesNotSameType typeThen typeElse); TA.ErrorExp ^! ERROR)
            ) else (Err.error err pos (EFmt.errorIntRequired typeTest); TA.ErrorExp ^! ERROR)
          )
    | A.WhileExp {test; body} -> (
        let transTest = (trexp test) in
        let typeTest = getType transTest in
      (match typeTest with 
      | INT -> (
        let transBody = trexp body in 
        let typeBody = getType transBody in
        (match typeBody with
        | VOID -> TA.WhileExp {test=transTest; body=transBody} ^! VOID
        | _ -> Err.error err pos (EFmt.errorWhileShouldBeVoid typeBody); TA.ErrorExp ^! ERROR))
      | _ -> Err.error err pos (EFmt.errorIntRequired typeTest); TA.ErrorExp ^! ERROR))
    
      | A.LetExp {decls; body} -> (
        let current_context = ref {err; venv} in
        let new_decls = ref [] in
        for i = 0 to (List.length decls)-1 do
          let (result_context, resultDecl) = transDecl !current_context (List.nth decls i) in
          current_context := result_context;
          new_decls := !new_decls @ [resultDecl];
        done;
        let transBody = transExp !current_context body in
        let bodyType = getType transBody in
        TA.LetExp { decls= !new_decls; body=transBody } ^! bodyType
        )

    | A.VarExp var -> trvar var

    | A.AssignExp {var; exp} -> (
      let transExp = trexp exp in 
      let expType = getType transExp in
      let varexp = trvar var in 
      (match varexp with 
        | TA.Exp { exp_base = TA.VarExp var; pos; ty } -> 
          if ty = expType 
            then TA.AssignExp { var=var; exp=transExp } ^! VOID
            else (Err.error err pos (EFmt.errorAssignmentTypesShouldMatch ty expType); TA.ErrorExp ^! ERROR) 
        | _ -> Err.error err pos "Tried to assign to something that is not a variable"; TA.ErrorExp ^! ERROR (* We should (probably) never enter this case *)
      )
    )

    (* the rest of the cases do not need handling in Sem0 / Assignment 3 *)
    | _ -> raise NotSem0
    
  (*trvar recurs over absyn.Var. When the venv needs to be updated the trexp function needs to call transExp rather than itself, to access this function*)
  and trvar (A.Var {var_base; pos}) = 
    (match var_base with 
    | A.SimpleVar symbol -> 
      let varLookup = (S.look (venv, symbol)) in 
      (match varLookup with 
      | Some (VarEntry ty) -> (
        let simple = TA.SimpleVar symbol in
        let var = TA.Var { var_base=simple; pos=pos; ty=ty } in 
        TA.Exp { exp_base = TA.VarExp var; pos=pos; ty=ty } )
        | None -> Err.error err pos (EFmt.errorVariableUndefined symbol); TA.Exp { exp_base = TA.ErrorExp; pos=pos; ty=ERROR }
        | Some((FunEntry {formals=_; result=_}) )-> Err.error err pos (EFmt.errorUsingFunctionAsVariable symbol); TA.Exp { exp_base = TA.ErrorExp; pos=pos; ty=ERROR }
      )
    | _ -> raise NotSem0
    )
  in
  trexp e

(*this function translates the declarations into new environments containing them, used in the letinexp body*)
and transDecl ({err; venv} : context) dec =
  match dec with
  | A.VarDec {name; escape; typ=Some (symbol, _); init; pos=pos2} -> (
    let transInit = transExp {err; venv} init in 
    let initType = getType transInit in
    let initTypeString = String.lowercase_ascii (PT.string_of_type initType) in
    (if initType = Ty.VOID || initType = Ty.ERROR || initType = Ty.NIL
      then (Err.error err pos2 (EFmt.errorVoid); {err; venv=(S.enter (venv, name, (VarEntry Ty.ERROR))) }, TA.VarDec { name; escape; typ=ERROR; init=transInit; pos=pos2 }) 
    else (
      if initTypeString = S.name symbol
      then 
        ({err; venv=(S.enter (venv, name, (VarEntry initType))) }, TA.VarDec { name; escape; typ=initType; init=transInit; pos=pos2 }) 
      else
        (Err.error err pos2 (EFmt.errorValDeclTypesShouldMatch initTypeString (S.name symbol)); {err; venv=(S.enter (venv, name, (VarEntry initType))) }, TA.VarDec { name; escape; typ=ERROR; init=transInit; pos=pos2 })))
      )
  | _ -> raise NotSem0

(* no need to change the implementation of the top level function *)
let transProg (e : A.exp) : TA.exp * Err.errenv =
  let err = Err.initial_env in
  try (transExp {venv= E.baseVenv; err} e, err)
  with NotSem0 ->
    Err.error err Lexing.dummy_pos
      "found language feature that is not part of sem0" ;
    ( TA.Exp
        { exp_base= TA.IntExp 0 (* dummy value *)
        ; pos= Lexing.dummy_pos
        ; ty= Ty.ERROR }
    , err )


(*When in doubt: append to error environment which is full of errors 
  which live in error state in united states of error (the best error 
  land in the error universe) land of the errors, home of the mistakes*)