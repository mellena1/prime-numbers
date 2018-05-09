package main

import (
	"fmt"
	"math"
	"sync"
	"time"
)

var numThreads = 8
var numPrimes = 1000000000

type numberLock struct {
	value *int
	lock  *sync.Cond
}

func main() {
	fmt.Printf("Looking through the first %d numbers for primes...\n", numPrimes)

	sTTime, sTPrimes := singlethread()
	fmt.Printf("Single thread: %v\n", sTTime)
	iPTime, iPPrimes := multithreadInnerParallel()
	fmt.Printf("Multi thread Inner Parallel: %v\n", iPTime)

	iPSame := true
	for i := 0; i < numPrimes; i++ {
		if !(sTPrimes[i] == iPPrimes[i]) {
			iPSame = false
		}
	}
	fmt.Printf("iP same: %v\n", iPSame)
}

func makePrimeArray() *[]bool {
	primes := make([]bool, numPrimes)
	primes[0] = true
	primes[1] = true // both 0 and 1 aren't prime
	return &primes
}

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

func multithreadInnerParallel() (time.Duration, []bool) {
	start := time.Now()
	primes := makePrimeArray()
	curNum := 2
	for curNum != -1 {
		sieveParallel(curNum, primes, 1000)
		curNum = nextNum(curNum, *primes)
	}
	return time.Since(start), *primes
}

func nextNum(curNum int, primes []bool) int {
	if curNum == -1 {
		return -1
	}
	var i int
	if curNum == 2 {
		i = curNum + 1
	} else {
		i = curNum + 2
	}
	for ; i <= numPrimes/2; i += 2 {
		if !primes[i] {
			return i
		}
	}
	return -1
}

func sievehelper(curNum int, primes *[]bool) {
	for i := curNum + curNum; i < numPrimes; i += curNum {
		(*primes)[i] = true
	}
}

func sieveParallel(curNum int, primes *[]bool, tasksPerThread int) {
	totalNumbers := numPrimes - 1 - curNum
	numbersToCheckOff := int(totalNumbers / curNum)
	if numbersToCheckOff < tasksPerThread {
		sievehelper(curNum, primes)
	} else {
		numberOfThreads := int(math.Ceil(float64(numbersToCheckOff) / float64(tasksPerThread)))
		step := tasksPerThread * curNum
		channels := make([]chan bool, numberOfThreads)
		for i := range channels {
			start := (step * i) + curNum + curNum
			end := (step * (i + 1)) + curNum
			if i == len(channels)-1 && numbersToCheckOff%tasksPerThread != 0 {
				end -= ((tasksPerThread - (numbersToCheckOff % tasksPerThread)) * curNum)
			}
			channels[i] = make(chan bool)
			go sievehelperParallel(channels[i], curNum, start, end, primes)
		}
		for _, ch := range channels {
			<-ch
		}
	}
}

func sievehelperParallel(ch chan bool, curNum int, start int, end int, primes *[]bool) {
	for i := start; i <= end; i += curNum {
		(*primes)[i] = true
	}
	ch <- true
}
