package ps

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var scriptsPath string

// Runner - contains specific info for PS script execution
type Runner struct {
	Cmd  string
	Args []string
	Ext  string
}

// Windows scpirts execution data
var Windows = Runner{
	Cmd:  "cscript.exe",
	Args: []string{"/nologo"},
	Ext:  ".vbs",
}

var std = Windows

func init() {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		log.Panic("Filepath not found!")
	}
	scriptsPath = filepath.Join(filepath.Dir(file), "scripts")
}

// PasteImage - pastes image to photoshop
func PasteImage(imgPath string, layerName string, x int, y int) ([]byte, error) {
	return ExecJsxScript(filepath.Join(scriptsPath, "pasteAndMove.jsx"), imgPath, layerName, strconv.FormatInt(int64(x+20), 10), strconv.FormatInt(int64(y), 10))
}

// ExecJsxScript - execute jsx script with given path
func ExecJsxScript(path string, args ...string) ([]byte, error) {
	return ExecScriptByName("execJs", append([]string{path}, args...)...)
}

//ExecScriptByName - execute script from "scripts" folder
func ExecScriptByName(prodName string, args ...string) ([]byte, error) {
	var res, errs bytes.Buffer
	cmd := exec.Command(std.Cmd, parseArgs(prodName, args)...)
	cmd.Stdout = &res
	cmd.Stderr = &errs
	if err := cmd.Run(); err != nil || len(errs.Bytes()) != 0 {
		return res.Bytes(), fmt.Errorf(`err: "%s"
		srrs.String(): "%s"
		args: "%s"
		res: "%s"`, err, errs.String(), args, res.String())
	}
	return res.Bytes(), nil
}

func parseArgs(name string, args []string) []string {
	if !strings.HasSuffix(name, std.Ext) {
		name += std.Ext
	}
	res := append(std.Args, filepath.Join(scriptsPath, name))

	if strings.Contains(name, "execJs") {
		res = append(res, args[0], fmt.Sprint(strings.Join(args[1:], "|")))
	} else {
		res = append(res, args...)
	}

	return res
}
