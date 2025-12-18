package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/VzoelFox/morphlang/pkg/agent"
)

func main() {
	record := flag.Bool("record", false, "Record a new interaction")
	user := flag.String("user", "", "User message")
	assistant := flag.String("assistant", "", "Assistant message")
	verify := flag.Bool("verify", false, "Verify memory integrity")

	flag.Parse()

	mem, err := agent.LoadMemory(agent.MemoryFile)
	if err != nil {
		fmt.Printf("Error loading memory: %v\n", err)
		os.Exit(1)
	}

	if *verify {
		fmt.Println("Integrity check passed.")
		fmt.Printf("Total interactions: %d\n", len(mem.Interactions))
		return
	}

	if *record {
		if *user == "" || *assistant == "" {
			fmt.Println("Please provide -user and -assistant messages")
			os.Exit(1)
		}
		if err := mem.Record(*user, *assistant); err != nil {
			fmt.Printf("Error recording: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Interaction recorded.")
	}
}
