/*
Copyright © 2023 Hantsaniala Eléo <hantsaniala@gmail.com>
*/
package main

import (
	"fmt"

	"github.com/hantsaniala/hStream/cmd"
)

const (
	banner = `
 _    ___ _
| |_ / __| |_ _ _ ___ __ _ _ __
| ' \\__ \  _| '_/ -_) _\ | '  \
|_||_|___/\__|_| \___\__,_|_|_|_|

`
)

func main() {
	fmt.Print(banner)
	cmd.Execute()
}
