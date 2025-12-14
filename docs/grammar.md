# Morph Language Grammar (Formal EBNF)

This document defines the formal grammar for the Morph programming language using Extended Backus-Naur Form (EBNF).
It serves as the **Single Source of Truth** for the Lexer and Parser implementation.

**Status:** ACTIVE
**Reference:** Based on `AGENTS.md` (v1.0.1)

---

## 1. Lexical Structure & Whitespace Rules

### 1.1 Strict Whitespace Rule (Zero Ambiguity)
Unlike many languages, Morph enforces strict whitespace rules for binary operators to ensure readability and prevent ambiguity.

*   **Binary Operators:** MUST be surrounded by whitespace.
    *   Valid: `a + b`, `x == 10`, `cnt += 1`
    *   Invalid: `a+b`, `x==10`, `cnt+=1`
*   **Unary Operators:** MUST NOT be followed by whitespace (preferred) or loosely allowed? `AGENTS.md` says `x = -5` and `x = - 5` are both Valid. But Binary MUST have spaces.
*   **Delimiters:** Parentheses `()`, Braces `{}`, Brackets `[]`, Comma `,` do not require strict spacing, but standard style applies.

### 1.2 Comments
*   Single line: Starts with `#` and extends to end of line.

---

## 2. Grammar Definition

```ebnf
/* Top Level */
program = { statement } ;

/* Statements */
statement =
    | return_statement
    | assignment_statement
    | if_expression       /* In Morph, if is an expression but can be used as statement */
    | while_expression
    | expression_statement
    ;

block = { statement } ;

/* Function Definition */
/* 'fungsi' can be a statement (declaration) or expression (literal) */
function_definition = "fungsi" , [ identifier ] , "(" , [ parameter_list ] , ")" , block , "akhir" ;

parameter_list = identifier , { "," , identifier } ;

/* Control Flow */
if_expression = "jika" , expression , block ,
    { "atau_jika" , expression , block } ,
    [ "lainnya" , block ] ,
    "akhir" ;

while_expression = "selama" , expression , block , "akhir" ;

return_statement = "kembalikan" , [ expression ] , [ ";" ] ;

/* Assignment & Variables */
/* Implicit declaration via assignment */
assignment_statement = identifier , "=" , expression , [ ";" ] ;

expression_statement = expression , [ ";" ] ;

/* Expressions (Precedence: Lowest to Highest) */
expression = logic_or ;

logic_or = logic_and , { "||" , logic_and } ;
logic_and = equality , { "&&" , equality } ;

equality = comparison , { ( "==" | "!=" ) , comparison } ;
comparison = term , { ( "<" | ">" | "<=" | ">=" ) , term } ;

term = factor , { ( "+" | "-" ) , factor } ;
factor = unary , { ( "*" | "/" | "%" ) , unary } ;

unary = ( "!" | "-" ) , unary | primary ;

primary =
    | integer_literal
    | boolean_literal
    | string_literal
    | identifier
    | function_definition
    | map_literal
    | "(" , expression , ")"
    | call_expression
    | index_expression /* map/array access */
    ;

/* Suffix Expressions (Call, Index) */
/* Note: Simplified EBNF. Real parser uses precedence for suffixes. */
call_expression = primary , "(" , [ argument_list ] , ")" ;
index_expression = primary , ( "." , identifier | "[" , expression , "]" ) ;

argument_list = expression , { "," , expression } ;

/* Literals */
integer_literal = digit , { digit } ;
boolean_literal = "benar" | "salah" ;
identifier = letter , { letter | digit | "_" } ;

/* String Interpolation */
string_literal = '"' , { string_content | interpolation } , '"' ;
interpolation = "#{" , expression , "}" ;
string_content = ? Any char except " and #{ ? ;

/* Map Literal */
map_literal = "{" , [ map_entries ] , "}" ;
map_entries = map_entry , { "," , map_entry } ;
map_entry = ( identifier | string_literal ) , ":" , expression ;

```

## 3. Operator Precedence

| Priority | Operator | Description | Associativity |
|----------|----------|-------------|---------------|
| 1 (Low)  | `||`     | Logical OR  | Left          |
| 2        | `&&`     | Logical AND | Left          |
| 3        | `==`, `!=` | Equality  | Left          |
| 4        | `<`, `>`, `<=`, `>=` | Comparison | Left |
| 5        | `+`, `-` | Sum         | Left          |
| 6        | `*`, `/`, `%` | Product | Left          |
| 7        | `-`, `!` | Unary       | Right         |
| 8        | `( )`    | Call        | Left          |
| 8        | `.` , `[ ]` | Access   | Left          |

## 4. Error Handling Specs

*   **Syntax Errors:** Must report exact Line and Column.
*   **Missing Whitespace:** If a binary operator is found without surrounding whitespace, report error `E001`.
    *   _Check:_ `Tokens adjacent to Binary Op`.
*   **Unclosed Blocks:** Must report missing `akhir`.
