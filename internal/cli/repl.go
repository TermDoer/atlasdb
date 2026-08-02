package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/TermDoer/atlasdb/internal/storage"
)
func Start() {
	scanner := bufio.NewScanner(os.Stdin)

	store := storage.NewMemeoryStore()
	for scanner.Err() == nil {
		fmt.Print("> ")

		if !scanner.Scan(){
			break
		}

		line := scanner.Text()

		if (strings.Contains(strings.ToUpper(line), "EXIT")) {
			break
		}
		cmd, ok := Parse(line)
		if ok == nil {
			Execute(cmd, store)
		} else {
			fmt.Println(ok)
		}
	}
}