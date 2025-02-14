package main

import (
	"bufio"
	"fmt"
	"log"
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

// Логирование входных данных
func inputSource(output chan int, done <-chan struct{}) {
	scanner := bufio.NewScanner(os.Stdin)
	log.Println("Источник данных запущен.")
	fmt.Println("Введите числа через пробел (exit для выхода):")
	for scanner.Scan() {
		select {
		case <-done:
			log.Println("Источник данных завершен.")
			close(output)
			return
		default:
			text := scanner.Text()
			if text == "exit" {
				log.Println("Пользователь запросил завершение.")
				close(output)
				return
			}
			numbers := strings.Fields(text)
			for _, numStr := range numbers {
				if num, err := strconv.Atoi(numStr); err == nil {
					output <- num
					log.Printf("Введено число: %d", num)
				} else {
					log.Printf("Некорректный ввод: %s", numStr)
					fmt.Printf("Некорректный ввод: %s\n", numStr)
				}
			}
		}
	}
	close(output)
	log.Println("Источник данных завершен.")
}

// Логирование фильтрации отрицательных чисел
func filterNegative(input, output chan int, done <-chan struct{}) {
	log.Println("Фильтр отрицательных чисел запущен.")
	for {
		select {
		case <-done:
			log.Println("Фильтр отрицательных чисел завершен.")
			close(output)
			return
		case num, ok := <-input:
			if !ok {
				log.Println("Фильтр отрицательных чисел завершен из-за закрытия канала.")
				close(output)
				return
			}
			if num >= 0 {
				output <- num
				log.Printf("Число прошло фильтр отрицательных: %d", num)
			} else {
				log.Printf("Число отфильтровано как отрицательное: %d", num)
			}
		}
	}
}

// Логирование фильтрации чисел, не кратных трем
func filterNonMultiplesOfThree(input, output chan int, done <-chan struct{}) {
	log.Println("Фильтр чисел, не кратных трем, запущен.")
	for {
		select {
		case <-done:
			log.Println("Фильтр чисел, не кратных трем, завершен.")
			close(output)
			return
		case num, ok := <-input:
			if !ok {
				log.Println("Фильтр чисел, не кратных трем, завершен из-за закрытия канала.")
				close(output)
				return
			}
			if num != 0 && num%3 == 0 {
				output <- num
				log.Printf("Число прошло фильтр кратности трем: %d", num)
			} else {
				log.Printf("Число отфильтровано как не кратное трем: %d", num)
			}
		}
	}
}

// Логирование буферизации
func bufferStage(input, output chan int, done <-chan struct{}) {
	buffer := NewRingBuffer(bufferSize)
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()
	log.Println("Буферизация запущена.")
	for {
		select {
		case <-done:
			log.Println("Буферизация завершена.")
			flushBuffer(buffer, output)
			close(output)
			return
		case num, ok := <-input:
			if !ok {
				log.Println("Буферизация завершена из-за закрытия канала.")
				flushBuffer(buffer, output)
				close(output)
				return
			}
			buffer.Add(num)
			log.Printf("Число добавлено в буфер: %d", num)
		case <-ticker.C:
			flushBuffer(buffer, output)
		}
	}
}

// Промежуточная отправка данных из буфера
func flushBuffer(buffer *RingBuffer, output chan int) {
	for _, val := range buffer.Flush() {
		output <- val
	}
}

// Логирование потребителя
func consumer(input chan int, done <-chan struct{}) {
	log.Println("Потребитель запущен.")
	for {
		select {
		case <-done:
			log.Println("Потребитель завершен.")
			return
		case num, ok := <-input:
			if !ok {
				log.Println("Потребитель завершен из-за закрытия канала.")
				return
			}
			log.Printf("Получены данные: %d", num)
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
	log.Println("Программа завершена.")
}
