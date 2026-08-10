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

	fmt.Print("1. Go map \n2. WAL \n3. SSTable \n4. LSM\n> ")
	
	var store storage.Storage
	if scanner.Scan(){
		line := scanner.Text()
	switch line {
		case "1":
			store = storage.NewMemeoryStore()
		case "2":
			store = storage.NewWalStore() 
		case "3":
			maxlen := 10
			store = storage.NewSSTableStore(maxlen)
		case "4":
			maxlen := 10
			store = storage.NewLSMStore(maxlen)
	}
	}
	
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