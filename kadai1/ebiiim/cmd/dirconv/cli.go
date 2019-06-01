package dirconv

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gopherdojo/dojo5/kadai1/ebiiim/cmd/conv"
)

const usageSrcExt = `source extension (jpg, png, tiff, bmp)`
const usageTgtExt = `target extension (jpg, png, tiff, bmp)`

// Usage string
const Usage = `Usage:
  imgconv [-source_ext=<ext>] [-target_ext=<ext>] DIR
Arguments:
  -source_ext=<ext>` + "\t" + usageSrcExt + ` [default: jpg]
  -target_ext=<ext>` + "\t" + usageTgtExt + ` [default: png]`

// Cli struct
type Cli struct {
	// directory name to traverse
	Dir string
	// source extension
	SrcExt string
	// target extension
	TgtExt string
}

// Result struct
type Result struct {
	// this value is usually not continuous because DirConv uses goroutine
	Index int
	// relative path from the dir passed to args
	RelPath string
	// if err == nil then true
	IsOk bool
}

// DirConv runs an imgconv command (parsed by NewCli()).
//   1. traverses dirs
//   2. converts files
//   3. shows logs and returns results
// Returns a list of results likes:
//   [{0 dummy.jpg false} {2 dirA/figB.jpg true} {1 figA.jpg true} ...]
func (cli Cli) DirConv() []Result {
	var results []Result

	// show help if no dir specified
	if cli.Dir == "" {
		fmt.Println(Usage)
		os.Exit(0)
	}

	// get file paths to convert
	files, err := cli.traverseImageFiles()
	if err != nil {
		panic(err)
	}

	// convert files (goroutined)
	wg := &sync.WaitGroup{}
	for i, v := range files {
		wg.Add(1)
		go func(idx int, val string) {
			defer wg.Done()
			oldFileName := fmt.Sprintf("%s/%s", cli.Dir, val)
			newFileName := oldFileName[0:len(oldFileName)-len(cli.SrcExt)] + cli.TgtExt
			log := fmt.Sprintf("%s -> %s", oldFileName, newFileName)

			ic := conv.ImgConv{SrcPath: oldFileName, TgtPath: newFileName}
			err := ic.Convert()

			ok := true
			if err != nil {
				ok = false
				_, _ = fmt.Fprintln(os.Stderr, err)
				log = fmt.Sprintf("[Failed] %s", log)
			} else {
				log = fmt.Sprintf("[OK] %s", log)
			}

			results = append(results, Result{Index: idx, RelPath: val, IsOk: ok})
			fmt.Println(log)
		}(i, v)
	}
	wg.Wait()

	return results
}

func (cli *Cli) traverseImageFiles() (files []string, err error) {
	err = filepath.Walk(cli.Dir,
		func(path string, info os.FileInfo, err error) error {
			relPath, err := filepath.Rel(cli.Dir, path)
			if !info.IsDir() && err == nil && strings.ToLower(filepath.Ext(relPath)) == cli.SrcExt {
				files = append(files, relPath)
			}
			return nil
		})
	return
}

// NewCli initializes a Cli struct with given args.
func NewCli(args []string) (cli *Cli) {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	var (
		srcExt = flags.String("source_ext", "jpg", usageSrcExt)
		tgtExt = flags.String("target_ext", "png", usageTgtExt)
	)
	err := flags.Parse(args[1:])
	if err != nil {
		panic(err)
	}
	dir := flags.Arg(0) // get the first dir name only

	formatExt(srcExt)
	formatExt(tgtExt)

	cli = &Cli{dir, *srcExt, *tgtExt}
	return
}

func formatExt(ext *string) {
	*ext = strings.ToLower(*ext)
	if !strings.HasPrefix(*ext, ".") {
		*ext = "." + *ext
	}
}
