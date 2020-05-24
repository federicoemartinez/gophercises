package main

import (
	"flag"
	"time"
)
import "os"
import "fmt"
import "bufio"
import "encoding/csv"
import "strings"

func main() {
	csvFileName := flag.String("csv", "problems.csv", "a csv file with questions")
	secondsTimeout := flag.Int("timeout", 30, "Timeout to answer a question. You lose if you take more time than that")
	flag.Parse()

	file, err := openCSV(csvFileName)
	defer file.Close()

	csvLines := getCSVLines(file, err)

	reader := bufio.NewReader(os.Stdin)

	// We will use a channel to read from stdin so i can do a select and stop the quiz game if the timer is fired
	inputChannel := make(chan string, 1)
	correctAnswers := 0
	wrongAnswers := 0

	println("Press enter to start")
	reader.ReadLine()

	for _, splittedLine := range csvLines {

		answer, timeout, text := processQuestion(splittedLine, secondsTimeout, reader, inputChannel)

		if timeout {
			break
		}
		if answer != text {
			wrongAnswers += 1
		} else {
			correctAnswers += 1
		}

	}

	fmt.Printf("Correct answers %d\n", correctAnswers)
	fmt.Printf("Total answers %d\n", correctAnswers+wrongAnswers)
}

func processQuestion(splittedLine []string, secondsTimeout *int, reader *bufio.Reader, inputChannel chan string) (string, bool, string) {
	question, answer := getQuestionAndAnswerFromLine(splittedLine)
	println(question)

	t := time.NewTimer(time.Duration(*secondsTimeout) * time.Second)

	go readInput(reader, inputChannel)
	timeout := false
	var text string
	select {
	case <-t.C:
		println("Time is up!")
		timeout = true
	case text = <-inputChannel:
		text = strings.ToUpper(strings.TrimSpace(text))
		t.Stop()
	}
	return answer, timeout, text
}

func getQuestionAndAnswerFromLine(splittedLine []string) (string, string) {
	if len(splittedLine) != 2 {
		panic(fmt.Sprintf("Invalid csv line %s", splittedLine))
	}
	question := strings.TrimSpace(splittedLine[0])
	answer := strings.ToUpper(strings.TrimSpace(splittedLine[1]))
	return question, answer
}

func getCSVLines(file *os.File, err error) [][]string {
	scanner := csv.NewReader(file)
	csvLines, err := scanner.ReadAll()
	if err != nil {
		panic(fmt.Sprintf("Error processing csv file %s", err.Error()))
	}
	return csvLines
}

func openCSV(csvFileName *string) (*os.File, error) {
	if csvFileName == nil {
		panic("csv argument is required")
	}
	file, err := os.Open(*csvFileName)
	if err != nil {
		panic(fmt.Sprintf("File %s cannot be opened", *csvFileName))
	}
	return file, err
}

func readInput(reader *bufio.Reader, inputChannel chan string) {
	answer, _ := reader.ReadString('\n')
	inputChannel <- answer
}
