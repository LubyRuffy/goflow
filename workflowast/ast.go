package workflowast

import parsec "github.com/prataprc/goparsec"

var (
	globalAst    *parsec.AST
	globalParser parsec.Parser
)

func init() {
	globalAst, globalParser = newAst()
}

// newAst 构建语法树
func newAst() (*parsec.AST, parsec.Parser) {
	ast := parsec.NewAST("program", 100)

	space := parsec.Token(`\s`, "SPACE")
	spaceMaybe := parsec.Maybe(nil, space)
	pipeOperator := parsec.Atom(`&`, "pipe_operator")
	openP := parsec.Atom(`(`, "OPENP")
	closeP := parsec.Atom(`)`, "CLOSEP")
	openFork := parsec.Atom(`[`, "OPENFORK")
	closeFork := parsec.Atom(`]`, "CLOSEFORK")
	andFork := parsec.Atom(`|`, "ANDFORK")
	comma := parsec.Atom(",", "COMMA")
	null := parsec.Atom("null", "null")
	boolean := parsec.OrdChoice(nil, parsec.Atom("true", "BOOL"), parsec.Atom("false", "BOOL"))
	doubleQuoteString := parsec.Token(`"(?:\\"|.)*?"`, "DOUBLEQUOTESTRING")
	quoteString := parsec.Token("`(?:\\`|.)*?`", "QUOTESTRING")

	var function parsec.Parser
	var fork parsec.Parser
	var pipe parsec.Parser

	// value 值表达式，可以是function
	identifier := parsec.OrdChoice(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			switch node := nodes[0].(type) {
			case []parsec.ParsecNode:
				switch n := node[0].(type) {
				case *parsec.Terminal:
					return n
				}
			}
			return nodes[0]
		},
		parsec.Float(), parsec.Hex(), parsec.Int(),
		parsec.Oct(),
		parsec.Char(),
		doubleQuoteString, quoteString, //parsec.String(),
		null, boolean,
		&function,
	)
	value := ast.And("value", nil, spaceMaybe, identifier, spaceMaybe)

	// 参数 Ast parameter list -> value | value "," value
	parameter := ast.Kleene("parameter", nil, value, comma)
	parameterList := ast.Maybe("parameterList", nil, parameter)

	// 函数
	function = ast.And("function", nil, spaceMaybe, parsec.Ident(), openP, spaceMaybe, parameterList, spaceMaybe, closeP, spaceMaybe)
	workflow := ast.OrdChoice("workflow", nil, &fork, function)

	// 分叉
	pipeList := ast.Kleene("pipeList", nil, &pipe, andFork)
	fork = ast.And("fork", nil, openFork, &pipeList, closeFork)

	// 最终的pipe
	pipe = ast.Kleene("pipe", nil, workflow, pipeOperator)

	return ast, pipe
}
