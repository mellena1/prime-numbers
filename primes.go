package main

import (
	"fmt"
	"time"
)

var numThreads = 8
var numPrimes = 1000000000

func main() {
	fmt.Printf("Looking through the first %d numbers for primes...\n", numPrimes)

	sTTime, sTPrimes := singlethread()
	fmt.Printf("Single thread: %v\n", sTTime)
	parallelTime, parallelPrimes := multithread()
	fmt.Printf("Multi thread Inner Parallel: %v\n", parallelTime)

	fmt.Printf("Parallel same result: %v\n", checkArrays(sTPrimes, parallelPrimes))
}

// checkArrays ... Return true if the two arrays have the same values
// Assumes that both arrays are size numPrimes
func checkArrays(array1, array2 []bool) bool {
	same := true
	for i := 0; i < numPrimes; i++ {
		if !(array1[i] == array2[i]) {
			same = false
			break
		}
	}
	return same
}

// makePrimeArray ... Make a prime array of size numPrimes
func makePrimeArray() *[]bool {
	primes := make([]bool, numPrimes)
	primes[0] = true
	primes[1] = true // both 0 and 1 aren't prime
	return &primes
}

// singlethread ... Find primes with main thread
func singlethread() (time.Duration, []bool) {
	start := time.Now()
	primes := makePrimeArray()
	curNum := 2
	for curNum != -1 {
		sievehelper(curNum, primes)
		curNum = nextNum(curNum, *primes)
	}
	return time.Since(start), *primes
}

// multithread ... Find primes with multiple threads
func multithread() (time.Duration, []bool) {
	start := time.Now()
	primes := makePrimeArray()
	curNum := 2
	for curNum != -1 {
		sieveParallel(curNum, primes)
		curNum = nextNum(curNum, *primes)
	}
	return time.Since(start), *primes
}

// nextNum ... Returns the next number that the sieve should be run against
func nextNum(curNum int, primes []bool) int {
	var i int
	// Always go to next odd number
	if curNum == 2 {
		i = 3
	} else {
		i = curNum + 2
	}
	// Check through primes for the next number that hasn't been marked not prime
	for ; i <= numPrimes/2; i += 2 {
		if !primes[i] {
			return i
		}
	}
	// If there aren't anymore, return -1
	return -1
}

// sievehelper ... Run the sieve single threaded
func sievehelper(curNum int, primes *[]bool) {
	// Mark 2*curNum, 3*curNum, 4*curNum, etc as not prime
	for i := curNum + curNum; i < numPrimes; i += curNum {
		(*primes)[i] = true
	}
}

// sieveParallel ... Runs the sieve parallely in numThreads for a given curNum
func sieveParallel(curNum int, primes *[]bool) {
	numbersToCheckOff := int(numPrimes / curNum)
	tasksPerThread := int(numbersToCheckOff / numThreads)
	if tasksPerThread < 100 { // Starting up threads aren't worth it for small num of tasks
		// Run single threaded in these cases
		sievehelper(curNum, primes)
	} else {
		step := tasksPerThread * curNum
		channels := make([]chan bool, numThreads)
		// Start up all of the goroutines
		for i := range channels {
			start := (step * i) + curNum + curNum
			// end will be the last index if last thread, else use steps to calculate it
			end := 0
			if i == numThreads-1 {
				end = numPrimes - 1
			} else {
				end = (step * (i + 1)) + curNum
			}
			// Start up the goroutine
			channels[i] = make(chan bool)
			go sievehelperParallel(channels[i], curNum, start, end, primes)
		}
		// Wait for all of the goroutines to be done
		for _, ch := range channels {
			<-ch
		}
	}
}

// sievehelperParallel ... function to be run parallely to mark off nonprimes
func sievehelperParallel(ch chan bool, curNum int, start int, end int, primes *[]bool) {
	for i := start; i <= end; i += curNum {
		(*primes)[i] = true
	}
	ch <- true
}
