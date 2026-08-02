package cli

import (
	"errors"
	"fmt"

	"github.com/TermDoer/atlasdb/internal/storage"
)

func Execute(cmd Command, store storage.Storage) error {
	if len(cmd.Args) < 1 {
		return errors.New("Key or Value missing")
	}
	switch cmd.Name {
	case "GET":
		value, ok := store.Get(cmd.Args[0])
		if ok == nil {
			fmt.Println(string(value))
		} else {
		fmt.Println(ok)
		}
		return ok
	case "SET":
		store.Set(cmd.Args[0], []byte(cmd.Args[1]))
		fmt.Println("ok")
		return nil
	case "DELETE":
		store.Delete(cmd.Args[0])
		fmt.Println("ok")
		return nil
	case "EXISTS":
		ok := store.Exists(cmd.Args[0])
		fmt.Println(ok)
		return nil
	}
	return nil
}