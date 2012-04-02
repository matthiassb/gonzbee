package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/DanielMorsing/gonzbee/nzb"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
)

var (
	profile   = flag.String("profile", "", "Where to save profile data")
	rm        = flag.Bool("rm", false, "Remove the nzb file after downloading")
	saveDir   = flag.String("d", "", "Save to this directory")
	aggregate = flag.String("a", "", "Save all files in all NZBs in this directory")
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(4)
	if *profile != "" {
		cpuprof := *profile + ".pprof"
		pfile, err := os.Create(cpuprof)
		if err != nil {
			panic(errors.New("Could not create profile file"))
		}
		defer pfile.Close()
		err = pprof.StartCPUProfile(pfile)
		if err != nil {
			panic(err)
		}
		defer pprof.StopCPUProfile()
	}

	fmt.Println(config)
	for _, path := range flag.Args() {
		file, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			continue
		}

		nzb, err := nzb.Parse(bytes.NewBuffer(file))
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			continue
		}

		var downloadDir string
		if *aggregate != "" {
			downloadDir = *aggregate
		} else {
			downloadDir = filepath.Base(path)
		}
		jobStart(nzb, downloadDir, *saveDir)
		if *rm {
			err = os.Remove(path)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
			}
		}
	}

	if *profile != "" {
		memprof := *profile + ".memprof"
		pfile, err := os.Create(memprof)
		if err != nil {
			panic(errors.New("Could not create profile file"))
		}
		defer pfile.Close()
		err = pprof.WriteHeapProfile(pfile)
		if err != nil {
			panic(err)
		}

		var memstat runtime.MemStats
		runtime.ReadMemStats(&memstat)
		b, err := json.MarshalIndent(memstat, "", "\t")
		if err != nil {
			panic(err)
		}
		err = ioutil.WriteFile(*profile+".memstats", b, 0644)
		if err != nil {
			panic(err)
		}
	}
}
