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
	channelsTime, channelsPrimes := multithreadChannels()
	fmt.Printf("Multi thread channels: %v\n", channelsTime)
	lockingTime, lockingPrimes := multithreadLocking()
	fmt.Printf("Multi thread locking: %v\n", lockingTime)
	goroutinesTime, goroutinesPrimes := multithreadGoRoutines()
	fmt.Printf("Multi thread goroutines: %v\n", goroutinesTime)

	iPSame := true
	channelsSame := true
	lockingSame := true
	goroutinesSame := true
	for i := 0; i < numPrimes; i++ {
		if !(sTPrimes[i] == iPPrimes[i]) {
			fmt.Printf("%v %v %v\n", i, sTPrimes[i], iPPrimes[i])
			iPSame = false
		}
		if !(sTPrimes[i] == channelsPrimes[i]) {
			channelsSame = false
		}
		if !(sTPrimes[i] == lockingPrimes[i]) {
			lockingSame = false
		}
		if !(sTPrimes[i] == goroutinesPrimes[i]) {
			goroutinesSame = false
		}
	}
	fmt.Printf("iP same: %v\nchannels same: %v\nlocking same: %v\ngoroutines same: %v\n", iPSame, channelsSame, lockingSame, goroutinesSame)
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

func multithreadChannels() (time.Duration, []bool) {
	start := time.Now()
	primes := makePrimeArray()
	channels := make([]chan int, numThreads)
	doneChannels := make([]chan bool, numThreads)
	quit := make([]bool, numThreads)
	for i := range channels {
		channels[i] = make(chan int)
		doneChannels[i] = make(chan bool)
	}
	curNum := 0
	firstPrimes := []int{2, 3, 5, 7, 11, 13, 17, 19}

	for i := range channels {
		go channelsSieve(channels[i], doneChannels[i], primes)
		if i < len(firstPrimes) {
			channels[i] <- firstPrimes[i]
			curNum = firstPrimes[i]
		} else {
			curNum = nextNum(curNum, *primes)
			channels[i] <- curNum
		}
	}

	for {
		allQuit := true
		for _, v := range quit {
			if !v {
				allQuit = false
				break
			}
		}
		if allQuit {
			break
		}
		for i, doneCh := range doneChannels {
			select {
			case <-doneCh:
				curNum = nextNum(curNum, *primes)
				if curNum == -1 {
					quit[i] = true
				}
				channels[i] <- curNum
			default:
				continue
			}
		}
	}
	return time.Since(start), *primes
}

func multithreadLocking() (time.Duration, []bool) {
	start := time.Now()
	primes := makePrimeArray()
	channels := make([]chan int, numThreads)
	numbers := make([]*numberLock, numThreads)
	doneChannels := make([]chan bool, numThreads)
	quit := make([]bool, numThreads)
	for i := range channels {
		channels[i] = make(chan int)
		doneChannels[i] = make(chan bool)
	}
	curNum := 0
	firstPrimes := []int{2, 3, 5, 7, 11, 13, 17, 19}

	for i := range channels {
		if i < len(firstPrimes) {
			numbers[i] = &numberLock{value: &firstPrimes[i], lock: sync.NewCond(&sync.Mutex{})}
			curNum = firstPrimes[i]
		} else {
			curNum = nextNum(curNum, *primes)
			number := curNum
			numbers[i] = &numberLock{value: &number, lock: sync.NewCond(&sync.Mutex{})}
		}
		go lockingSieve(numbers[i], primes)
	}

	for {
		allQuit := true
		for _, v := range quit {
			if !v {
				allQuit = false
				break
			}
		}
		if allQuit {
			break
		}

		for i := range numbers {
			if quit[i] {
				continue
			}
			if *numbers[i].value == -2 {
				curNum = nextNum(curNum, *primes)
				if curNum == -1 {
					quit[i] = true
				}
				numbers[i].lock.L.Lock()
				*numbers[i].value = curNum
				numbers[i].lock.Signal()
				numbers[i].lock.L.Unlock()
			}
		}
	}
	return time.Since(start), *primes
}

func multithreadGoRoutines() (time.Duration, []bool) {
	start := time.Now()
	primes := makePrimeArray()
	doneChannels := make([]chan bool, numThreads)
	for i := range doneChannels {
		doneChannels[i] = make(chan bool)
	}
	curNum := 0
	firstPrimes := []int{2, 3, 5, 7, 11, 13, 17, 19}

	for i, ch := range doneChannels {
		if i < len(firstPrimes) {
			curNum = firstPrimes[i]
		} else {
			curNum = nextNum(curNum, *primes)
		}
		go goRoutinesSieve(ch, curNum, primes)
	}

	allNil := true
	for {
		allNil = true
		for i := range doneChannels {
			if doneChannels[i] == nil {
				continue
			}
			allNil = false
			select {
			case <-doneChannels[i]:
				curNum = nextNum(curNum, *primes)
				if curNum == -1 {
					doneChannels[i] = nil
				} else {
					go goRoutinesSieve(doneChannels[i], curNum, primes)
				}
			default:
				continue
			}
		}
		if allNil {
			break
		}
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

func channelsSieve(ch chan int, done chan bool, primes *[]bool) {
	for {
		curNum := <-ch
		if curNum == -1 {
			return
		}
		sievehelper(curNum, primes)
		done <- true
	}
}

func lockingSieve(number *numberLock, primes *[]bool) {
	for {
		number.lock.L.Lock()
		for *number.value == -2 {
			number.lock.Wait()
		}
		number.lock.L.Unlock()
		if *number.value == -1 {
			return
		}
		sievehelper(*number.value, primes)
		*number.value = -2
	}
}

func goRoutinesSieve(done chan bool, number int, primes *[]bool) {
	sievehelper(number, primes)
	done <- true
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
