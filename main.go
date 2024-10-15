package simpl

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:]

	sourceCodeFile := args[0]
	_, err := os.ReadFile(sourceCodeFile)
	if err != nil {
		panic(fmt.Sprintf("Can't read source code file. Error: %v", err))
	}
}
