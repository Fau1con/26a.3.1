package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	bufferSize    = 5
	flushInterval = 2 * time.Second
)

type RingBuffer struct {
	data  []int
	size  int
	head  int
	tail  int
	count int
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		data: make([]int, size),
		size: size,
	}
}

func (rb *RingBuffer) Add(value int) {
	if rb.count == rb.size {
		rb.head = (rb.head + 1) % rb.size
	} else {
		rb.count++

	}
	rb.data[rb.tail] = value
	rb.tail = (rb.tail + 1) % rb.size
}

func (rb *RingBuffer) Flush() []int {
	if rb.count == 0 {
		return nil
	}
	var result []int
	for rb.count > 0 {
		result = append(result, rb.data[rb.head])
		rb.head = (rb.head + 1) % rb.size
		rb.count--
	}
	return result
}

func inputSource(output chan int, done <-chan struct{}) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Введите числа через пробел (exit для выхода):")
	for scanner.Scan() {
		select {
		case <-done:
			close(output)
			return
		default:
			text := scanner.Text()
			if text == "exit" {
				close(output)
				return
			}
			numbers := strings.Fields(text)
			for _, numStr := range numbers {
				if num, err := strconv.Atoi(numStr); err == nil {
					output <- num
				} else {
					fmt.Printf("Некорректный ввод: %s\n", numStr)
				}
			}
		}
	}
	close(output)
}

func filterNegative(input, output chan int, done <-chan struct{}) {
	for {
		select {
		case <-done:
			close(output)
			return
		case num, ok := <-input:
			if !ok {
				close(output)
				return
			}
			if num >= 0 {
				output <- num
			}
		}
	}
}

func filterNonMultiplesOfThree(input, output chan int, done <-chan struct{}) {
	for {
		select {
		case <-done:
			close(output)
			return
		case num, ok := <-input:
			if !ok {
				close(output)
				return
			}
			if num != 0 && num%3 == 0 {
				output <- num
			}
		}
	}
}

func bufferStage(input, output chan int, done <-chan struct{}) {
	buffer := NewRingBuffer(bufferSize)
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			flushBuffer(buffer, output)
			close(output)
			return
		case num, ok := <-input:
			if !ok {
				flushBuffer(buffer, output)
				close(output)
				return
			}
			buffer.Add(num)
		case <-ticker.C:
			flushBuffer(buffer, output)
		}
	}
}

func flushBuffer(buffer *RingBuffer, output chan int) {
	for _, val := range buffer.Flush() {
		output <- val
	}
}

func consumer(input chan int, done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		case num, ok := <-input:
			if !ok {
				return
			}
			fmt.Printf("Получены данные: %d\n", num)
		}
	}
}

func main() {
	inputChan := make(chan int)
	filterNegChan := make(chan int)
	filterThreeChan := make(chan int)
	bufferChan := make(chan int)
	done := make(chan struct{})

	go inputSource(inputChan, done)
	go filterNegative(inputChan, filterNegChan, done)
	go filterNonMultiplesOfThree(filterNegChan, filterThreeChan, done)
	go bufferStage(filterThreeChan, bufferChan, done)

	go func() {
		consumer(bufferChan, done)
		close(done)
	}()

	<-done
	fmt.Println("Программа завершена.")
}
