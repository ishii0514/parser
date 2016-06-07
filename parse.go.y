%{
package parse

import(
    "fmt"
    )


type Expression interface{}
type Token struct {
    token int
    literal string
}

type NumExpr struct {
    literal string
}
type BinOpExpr struct {
    left Expression
    operator rune
    right Expression
    }
%}

%union{
    token Token
    expr Expression
}

%type<expr> program
%type<expr> expr
%token<token> NUMBER SELECT

%left '+'

%%

program
    : expr
    {
        $$ = $1
        yylex.(*LexerWrapper).result = $$
    }

expr
    : NUMBER
    {
        $$ = NumExpr{literal: $1.literal}
     }
     | expr '+' expr
     {
        $$ = BinOpExpr{left: $1, operator: '+', right: $3}
     } 
%%

type LexerWrapper struct {
    l *lexer
    result Expression
}

func(w *LexerWrapper) Lex(lval *yySymType) int {
    item := w.l.nextItem()
    token := int(item.typ)
    //lexのシンボルとparserのシンボルの合わせ方
    //案1.直接同じものを使う。シングルバイト1文字はそのままintに変換。シンプルだけどparser側の宣言に引っ張られる。
    //案2.変換関数を用意する。lexerとparserが分離しやすい。
    //案1を採用。変換関数を都度追加するのが面倒そうなので。
    if item.typ == itemEOF {
        token = 0
    }
    lval.token = Token{token:token, literal: item.val}
    return token
}

func (w *LexerWrapper) Error(e string) {
    panic(e)
}

func Parse(input string) {
	w := LexerWrapper{l: lex(input)}
	yyParse(&w)
	fmt.Printf("%#v\n", w.result)
}
