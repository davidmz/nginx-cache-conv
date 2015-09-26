package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/jawher/mow.cli"
)

func main() {
	app := cli.App("ng-cache-conv", "nginx cache files converter")

	app.Command("file", "Convert single file and pass result to stdout", func(cmd *cli.Cmd) {
		fileName := cmd.StringArg("FILE", "", "old version cache file")
		cmd.Action = func() {
			f, err := os.Open(*fileName)
			if err != nil {
				fail("Cannot open file:", err)
			}
			defer f.Close()

			version := getVersion(f)
			if version == 3 {
				// no need to convertation
				io.Copy(os.Stdout, f)
			} else if version == 0 {
				if err := convertFile(f, os.Stdout); err != nil {
					fail("Conversion error:", err)
				}
			} else {
				fail("Unsupported file version:", version)
			}
		}
	})

	app.Command("file-version", "Print version of single cache file", func(cmd *cli.Cmd) {
		fileName := cmd.StringArg("FILE", "", "cache file")
		justVersion := cmd.BoolOpt("s short", false, "print only version number")
		cmd.Action = func() {
			f, err := os.Open(*fileName)
			if err != nil {
				fail("Cannot open file:", err)
			}
			defer f.Close()

			version := getVersion(f)
			if *justVersion {
				fmt.Println(version)
			} else {
				fmt.Println("File version:", version)
			}
		}
	})

	app.Command("stat", "Collect statistics of file versions in directory", func(cmd *cli.Cmd) {
		dirName := cmd.StringArg("DIR", "", "directory name")
		cmd.Action = func() {
			nFiles := 0
			versions := [MAX_VERSION + 1]int{}

			t0 := time.Now()
			go func() {
				c := time.Tick(1 * time.Second)
				for range c {
					fmt.Println("Time spent:", time.Since(t0))
					fmt.Println("Files processed:", nFiles)
					fmt.Println("Versions stats:")
					for i, v := range versions {
						fmt.Printf(" %d:\t%d\n", i, v)
					}
					fmt.Println("---")
				}
			}()

			err := filepath.Walk(*dirName, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() {
					nFiles++
					return func() error {
						f, err := os.Open(path)
						if err != nil {
							return err
						}
						defer f.Close()

						version := getVersion(f)
						if version <= MAX_VERSION {
							versions[version]++
						}

						return nil
					}()
				}
				return nil
			})
			if err != nil {
				fail("Cannot walk:", err)
			}
			fmt.Println("Time spent:", time.Since(t0))
			fmt.Println("Files processed:", nFiles)
			fmt.Println("Versions stats:")
			for i, v := range versions {
				fmt.Printf(" %d:\t%d (%d%%)\n", i, v, v*100/nFiles)
			}
		}
	})

	app.Run(os.Args)
}

func fail(msg ...interface{}) {
	fmt.Fprintln(os.Stderr, msg...)
	os.Exit(1)
}
