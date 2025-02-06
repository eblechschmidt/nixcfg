package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

// import "os"

// import "bytes"

func main() {
	cmd := exec.Command("nix", "repl", "--expr", "builtins.getFlake \"/home/eike/repos/nixos\"")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	var wg sync.WaitGroup

	evalOption := make(chan string)
	wg.Add(1)
	go func() {
		defer stdin.Close()
		defer wg.Done()

		for o := range evalOption {
			fmt.Printf("Evaluating: %s\n", o)
			io.WriteString(stdin, fmt.Sprintf("%s\n", o))
		}
	}()

	res := make(chan string)
	wg.Add(1)
	go func() {
		defer stdout.Close()
		defer wg.Done()
		// copy the data written to the PipeReader via the cmd to stdout
		// if _, err := io.Copy(os.Stdout, stdout); err != nil {
		//     log.Fatal(err)
		// }
		s := bufio.NewScanner(stdout)
		var sb strings.Builder
		multiline := false
		for {
			for s.Scan() {
				if s.Text() == "" {
					continue
				}
				if !multiline {
					if strings.HasPrefix(s.Text(), "{") {
						multiline = true
						sb.Reset()
					}
				}
				if multiline {
					sb.WriteString(fmt.Sprintf("%s\n", s.Text()))
					if strings.HasPrefix(s.Text(), "}") {
						res <- sb.String()
						multiline = false
						break
					}
					continue
				}
				res <- s.Text()
				multiline = false
			}
			if s.Err() != nil {
				break
			}
		}
		fmt.Println("Done reading")
	}()

	fmt.Println("Start process")
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	evalOption <- "nixosConfigurations.nixserve.config.snapraid.contentFiles"
	// <-res
	fmt.Println(<-res)
	evalOption <- "nixosConfigurations.nixserve.config.stylix\n"
	fmt.Println(<-res)
	// <-res
	close(evalOption)
	stdout.Close()
	cmd.Wait()
	wg.Wait()
}
