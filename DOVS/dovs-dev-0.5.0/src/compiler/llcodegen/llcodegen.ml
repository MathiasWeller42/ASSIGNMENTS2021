(**************************************************************************)
(* AU compilation.                                                        *)
(* Skeleton file -- expected to be modified as part of the assignment     *)
(**************************************************************************)

open Tigercommon
open Tigerhoist

(* Module aliases *)
module H = Habsyn
module S = Symbol
module Ty = Types
module B = Cfgbuilder

exception NotImplemented (* the final code should compile without this exception *)

exception Impossible (*if we get a mistake that should have been caught by semantic check*)

exception NotLLVM0

(* --- Helper functions and declarations --- *)
let is_user_visible fun1 = match fun1 with 
| "stringEqual" | "stringNotEq"| "stringLess"| "stringLessEq"|  "stringGreater" |  "stringGreaterEq" | "exponent" -> false
| _ -> true

let ( <$> ) f g x = f (g x) (* function application *)
let ( $> ) b1 b2 = b2 <$> b1 (* convenient for sequencing buildlets *)
let ty_of_exp (H.Exp {ty; _}) = ty (* type extractors *)
let ptr_i8 = Ll.Ptr Ll.I8 (* generic pointer type *)

(** [fresh s] generates fresh symbol starting with prefix [s] *)
let fresh : string -> S.symbol =
  let open Freshsymbols in
  let env = empty in
  gensym env

(** [aiws s i] adds instruction [i] with a fresh symbol starting with prefix
    [s] *)
let aiwf s i =
  let t = fresh s in
  (B.add_insn (Some t, i), Ll.Id t)

(* --- end of helper functions --- *)

(* Mapping Tiger built-in types to LLVM types *)
let ty_to_llty = function
  | Ty.INT -> Ll.I64
  | Ty.STRING -> ptr_i8
  | Ty.VOID -> Ll.Void
  | _ -> raise NotLLVM0

(* Suggested structure of the context for the translation *)
type context = {gdecls: (Ll.gid * Ll.gdecl) list ref; locals: Ll.ty S.table}

(* --- Main workhorse functions of the code generation --- *)

(** [cgExp ctxt exp] returns a pair of a buildlet and an operand *)
let rec cgExp ({gdecls; locals} : context) (Exp {exp_base; pos; ty} : H.exp) :
    B.buildlet * Ll.operand =
  let cgE = cgExp {gdecls; locals} in
  (* for recursive calls with the same context *)
  let open Ll in
  match exp_base with
  | H.IntExp n -> (B.id_buildlet, Ll.Const n)
  | H.StringExp s -> 
    let global_id = fresh "string" in
    let id_cast = fresh "string" in
    let s_length = String.length s in
    let struct_type = Ll.Struct [Ll.I64; Ll.Array (s_length, Ll.I8) ] in
    let struct_init = Ll.GStruct [(Ll.I64, Ll.GInt s_length); (Ll.Array(s_length, Ll.I8), GString s)] in
    let gdecl = (struct_type, struct_init) in
    gdecls := (global_id, gdecl) :: !gdecls;
    let result_block = B.add_insn(Some(id_cast), Ll.Bitcast(Ll.Ptr struct_type, Ll.Gid global_id, ptr_i8)) in
    (result_block, Ll.Id id_cast)

  | H.OpExp {left; oper; right} ->
      let build_right, op_right = cgE right in
      let build_left, op_left = cgE left in

      let exp_ty = ty_of_exp left in 

      (match exp_ty with 
      | Ty.INT -> (let bop =
        (match oper with 
        | PlusOp -> Some Ll.Add 
        | MinusOp -> Some Ll.Sub
        | TimesOp -> Some Ll.Mul
        | DivideOp -> Some Ll.SDiv
        | _ -> None) in

      let cnd = 
        (match oper with 
        | EqOp -> Some Ll.Eq
        | NeqOp -> Some Ll.Ne  
        | LtOp -> Some Ll.Slt
        | LeOp -> Some Ll.Sle
        | GtOp -> Some Ll.Sgt
        | GeOp -> Some Ll.Sge
        | _ -> None) in
        
        (match bop with 
        | Some (binop) -> (let i = Ll.Binop (binop, Ll.I64, op_left, op_right) in
        let newid = fresh "temp" in
        let b_insn = B.add_insn (Some newid, i) in
        let b_binop = B.seq_buildlets [build_left; build_right; b_insn] in
        (b_binop, Ll.Id newid))
        | None -> 
          ((match cnd with 
            | Some (cmp) ->
                (let i = Ll.Icmp (cmp, Ll.I64, op_left, op_right) in
                let newid = fresh "temp" in
                let b_insn = B.add_insn (Some newid, i) in
                let newid2 = fresh "temp" in
                let b_insn2 = B.add_insn (Some newid2, Ll.Zext(Ll.I1, Ll.Id newid, Ll.I64)) in
                let b_cmp = B.seq_buildlets [build_left; build_right; b_insn; b_insn2] in
                (b_cmp, Ll.Id newid2))
            | None -> (*this means we have an ExponentOp, so transform to a function call*) 
              let args_list = [left; right] in
              let exp = H.Exp {exp_base=H.CallExp { func=(S.symbol "exponent"); lvl_diff=0; args=args_list }; pos=pos; ty=ty} in
              cgE exp
        ))))
      | Ty.STRING -> 
        (
          let cnd_string = 
            (match oper with 
            | EqOp -> "stringEqual"
            | NeqOp -> "stringNotEq"
            | LtOp -> "stringLess"
            | LeOp -> "stringLessEq"
            | GtOp -> "stringGreater"
            | GeOp -> "stringGreaterEq"
            | _ -> raise Impossible) in
            
            let args_list = [left; right] in
            let exp = H.Exp {exp_base=H.CallExp { func=(S.symbol cnd_string); lvl_diff=0; args=args_list }; pos=pos; ty=ty} in
            cgE exp
        
        )
      | _ -> raise Impossible 
      )

      
      
  | H.CallExp {func; lvl_diff= 0; args} ->
      (* lvl_diff returned from the hoisting phase for Tiger Cub is always zero *)
      let mapped = List.map cgE args in 
      let mapped_builds, mapped_ops = List.split mapped in
      let types = List.map ty_of_exp args in
      let ll_types = List.map ty_to_llty types in
      let ops_types = ref (List.combine ll_types mapped_ops) in
      let dummy_ty = ptr_i8 in
      (if (is_user_visible (S.name func)) then ops_types := (dummy_ty, Ll.Null) :: !ops_types);
      let ret_type = ty_to_llty ty in 
      let ret_register = (match ret_type with 
        | Ll.Void -> None
        | _ -> Some (fresh "temp")
      ) in
      let call_block = B.add_insn(ret_register, Ll.Call(ret_type, Ll.Gid func, !ops_types)) in
      let result_block = B.seq_buildlets (List.append mapped_builds [call_block]) in
      let ret_operand = (match ret_register with 
        | Some (id) -> Ll.Id id
        | None -> Null
      ) in
      (result_block, ret_operand)

  | H.SeqExp exps -> 
      let mapped = List.map cgE exps in
      let mapped_builds = List.map fst mapped in
      let result_block = B.seq_buildlets mapped_builds in 
      let result_op = if (List.length mapped > 0) then snd (List.nth mapped ((List.length mapped) - 1)) else Null in
      (result_block, result_op)

  | H.IfExp {test; thn; els= Some e} -> 
      let build_test, op_test = cgE test in 
      let build_thn, op_thn = cgE thn in
      let build_els, op_els = cgE e in
      let common_var = fresh "temp" in
      let cast_test = fresh "temp" in
      let result = fresh "temp" in
      let thn_lbl = fresh "lbl" in
      let els_lbl = fresh "lbl" in
      let end_lbl = fresh "lbl" in
      let ret_ty = ty_to_llty ty in
      let test_with_term = (B.add_alloca(common_var, ret_ty) $> build_test $> B.add_insn(Some(cast_test), Ll.Icmp(Ll.Ne, Ll.I64, op_test, Ll.Const 0)) $>(B.term_block (Ll.Cbr (Ll.Id cast_test, thn_lbl, els_lbl)))) in
      let then_with_term = (B.start_block (thn_lbl)) $> build_thn $> B.add_insn (None, Ll.Store (ret_ty, op_thn, Ll.Id common_var)) $> B.term_block(Ll.Br end_lbl ) in
      let els_with_term  = (B.start_block (els_lbl)) $> build_els $> B.add_insn (None, Ll.Store (ret_ty, op_els, Ll.Id common_var)) $> B.term_block(Ll.Br end_lbl ) in
      let end_block = (B.start_block (end_lbl)) $> B.add_insn (Some(result), Ll.Load (ret_ty, Ll.Id common_var)) in
      let return_block = B.seq_buildlets [ test_with_term; then_with_term; els_with_term; end_block ] in
      (return_block, Ll.Id result)

  | H.WhileExp {test; body} -> 
      let test_lbl = fresh "lbl" in
      let body_lbl = fresh "lbl" in 
      let end_lbl = fresh "lbl" in
      let build_test, op_test = cgE test in
      let build_body, _ = cgE body in
      let cast_test = fresh "temp" in
      let start_block = B.term_block(Ll.Br test_lbl) in
      let test_with_term = (B.start_block (test_lbl)) $> build_test $> B.add_insn(Some(cast_test), Ll.Icmp(Ll.Ne, Ll.I64, op_test, Ll.Const 0)) $> (B.term_block (Ll.Cbr (Ll.Id cast_test, body_lbl, end_lbl))) in
      let body_with_term = (B.start_block (body_lbl)) $> build_body $> B.term_block(Ll.Br test_lbl) in
      let end_block = (B.start_block (end_lbl)) in 
      let result_block = B.seq_buildlets [ start_block; test_with_term; body_with_term; end_block ] in 
      (result_block, Null)

  | H.LetExp {vardecl= VarDec {name; typ; init; _}; body; _} ->
      let build_init, op_init = cgE init in
      let build_body, op_body = cgE body in
      let var_type = ty_to_llty typ in
      let var_block = B.add_alloca(name, var_type) $> B.add_insn(None, Ll.Store(var_type, op_init, Ll.Id name)) in
      let return_block = B.seq_buildlets [ build_init; var_block; build_body ] in
      (return_block, op_body)

  | H.VarExp v -> cgVar {gdecls; locals} v
  | H.AssignExp
      {var= Var {var_base= H.AccessVar (0, varname); ty= varty; _}; exp} ->
      (* first argument of the AccessVar is always zero in Tiger Cub *)
        let build_exp, op_exp = cgE exp in
        let store_block = B.add_insn(None, Ll.Store(ty_to_llty varty, op_exp, Ll.Id varname)) in
        let result_block = B.seq_buildlets [ build_exp ; store_block ] in
        (result_block, Null)
  (* the rest of the cases do not need handling in LLVM0/ Assignment 4 *)
  | _ -> raise NotLLVM0

and cgVar (ctxt : context) (H.Var {var_base; pos=_; ty}) =
  match var_base with
  | H.AccessVar (0, varname) -> 
      let var_loaded = fresh "temp" in
      let load_block = B.add_insn(Some(var_loaded), Ll.Load(ty_to_llty ty, Ll.Id varname)) in
      (load_block, Ll.Id var_loaded)
      (* first argument of the AccessVar is always zero in Tiger Cub *)
  | _ -> raise NotLLVM0

(** [cgTigerMain] returns a triple of the form (gdecls, llty, cfg) that
    corresponds to the global declarations, the llvm return type of this
    function, and the CFG of the main body *)
let cgTigerMain ty body locals =
  (* TODO: rewrite this function to include the following
          1) allocation of the locals
          2) call to cgExp with appropriate initalization of the context
          3) code generation of return from the function, and
          4) generating the final CFG and all gdecls
  *)
  let local_env = List.fold_left (fun env (symbol, ty) -> S.enter (env, symbol, (ty_to_llty ty))) S.empty locals in
  let ctxt = {gdecls= ref []; locals= local_env} in
  let build_body, op = cgExp ctxt body in
  let tr =
    match ty with
    | Ty.INT -> Ll.Ret (Ll.I64, Some op)
    | Ty.VOID -> Ll.Ret (Ll.Void, None)
    | Ty.STRING -> Ll.Ret(ptr_i8, Some op)
    | _ -> raise NotLLVM0
  in
  let tigermain_builder = B.seq_buildlets [build_body; B.term_block tr] in
  let cfg = tigermain_builder B.empty_cfg_builder |> B.get_cfg in
  (* obs: ocaml pipe operator |> *)
  (!(ctxt.gdecls), ty_to_llty ty, cfg)

(* --- No changes needed in the code below --- *)

(* For the starting assignment observe how the pattern matching expects there
   to be a function that is expected to be "tigermain" generated by the
   hoisting translation *)

let codegen_prog = function
  | H.Program
      { tdecls= []
      ; fdecls=
          [ H.Fdecl
              { name= fname
              ; args= []
              ; result
              ; body
              ; parent_opt= None
              ; locals
              ; _ } ] }
    when S.name fname = "tigermain" ->
      let gdecls, ret_ll_ty, tigermain_cfg =
        cgTigerMain result body locals
      in
      let open Ll in
      { tdecls= []
      ; gdecls
      ; fdecls=
          [ ( fname
            , { fty= ([ptr_i8], ret_ll_ty)
              ; param= [S.symbol "dummy"]
              ; cfg= tigermain_cfg } ) ] }
  | _ -> raise NotLLVM0
