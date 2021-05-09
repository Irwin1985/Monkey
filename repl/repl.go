package repl

import (
	"Monkey/evaluator"
	"bufio"
	"fmt"
	"io"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
)

// PROMPT es una constante que imprime las comillas en la consola.
const PROMPT = ">> "

// Start inicio de la consola REPL
func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	// constants := []object.Object{}
	// globals := make([]object.Object, vm.GlobalsSize)
	// symbolTable := compiler.NewSymbolTable()

	env := object.NewEnvironment()

	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParseErrors(out, p.Errors())
			continue
		}
		// inicio virtual machine
		// comp := compiler.NewWithState(symbolTable, constants)
		// err := comp.Compile(program)
		// if err != nil {
		// 	fmt.Fprintf(out, "Woops! Compilation failer:\n %s\n", err)
		// 	continue
		// }

		// code := comp.Bytecode()
		// constants = code.Constants

		// machine := vm.NewWithGlobalsStore(code, globals)
		// err = machine.Run()
		// if err != nil {
		// 	fmt.Fprintf(out, "Woops! Executing bytecode failed:\n %s\n", err)
		// 	continue
		// }

		// lastPopped := machine.LastPoppedStackElem()
		// io.WriteString(out, lastPopped.Inspect())
		// io.WriteString(out, "\n")
		// fin virtual machine

		evaluated := evaluator.Eval(program, env)
		if evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
	}
}

func printParseErrors(out io.Writer, errors []string) {
	io.WriteString(out, MONKEY_FACE)
	io.WriteString(out, "Woops! We ran into some monkey business here!\n")
	io.WriteString(out, " parse errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}

const MONKEY_FACE = `            __,__
   .--.  .-"     "-.  .--.
  / .. \/  .-. .-.  \/ .. \
 | |  '|  /   Y   \  |'  | |
 | \   \  \ 0 | 0 /  /   / |
  \ '- ,\.-"""""""-./, -' /
   ''-' /_   ^ ^   _\ '-''
       |  \._   _./  |
       \   \ '~' /   /
        '._ '-=-' _.'
           '-----'
`
