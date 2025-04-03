package main

import (
	"fmt"
	"os"
	"syscall"
)

func main() {
	f, err := os.OpenFile("lockfile.txt", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		fmt.Println("Could not acquire lock (another process has it)")
		return
	}

	fmt.Println("Lock acquired! Holding for a bit...")
	fmt.Println("Press Enter to release lock.")
	fmt.Scanln()

	syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	fmt.Println("Lock released.")
}
