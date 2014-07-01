package main

import (
	"fmt"
	"github.com/glendc/cgreader"
	"strings"
)

var rawProgramStream []byte
var currentStreamIndex int
var lineCounter, characterCounter, startLoopCounter, stopLoopCounter uint64
var streamIsValid bool

func RecursiveParser(command *Command) {
	for ; streamIsValid && currentStreamIndex < len(rawProgramStream); currentStreamIndex++ {
		characterCounter++

		switch rawProgramStream[currentStreamIndex] {
		case PI:
			command.add(&Command{func([]*Command) {
				programIndex++
			}, nil})

		case PD:
			command.add(&Command{func([]*Command) {
				programIndex--
			}, nil})

		case VI:
			command.add(&Command{func([]*Command) {
				programBuffer[programIndex]++
			}, nil})

		case VD:
			command.add(&Command{func([]*Command) {
				programBuffer[programIndex]--
			}, nil})

		case IN:
			command.add(&Command{func([]*Command) {
				if inputIsAvailable {
					if len(programInput) == 0 {
						programInput = []byte(<-inputChannel)
					}

					programBuffer[programIndex] = brainfuck_t(programInput[0])
					programInput = programInput[1:]
				}
			}, nil})

		case NOUT:
			command.add(&Command{func([]*Command) {
				if outputIsAvailable {
					outputChannel <- fmt.Sprintf("%d", programBuffer[programIndex])
				}
			}, nil})

		case COUT:
			command.add(&Command{func([]*Command) {
				if outputIsAvailable {
					outputChannel <- fmt.Sprintf("%s", string(programBuffer[programIndex]))
				}
			}, nil})

		case START:
			startLoopCounter++
			currentStreamIndex++

			loop := CreateLoopGroup()
			RecursiveParser(loop)
			command.add(loop)

		case STOP:
			if stopLoopCounter > startLoopCounter {
				fmt.Printf("ERROR! Parsing failed on Line %d (%d): encountered \"]\" while expecting ><+-,.#[\n", lineCounter, characterCounter)
				streamIsValid = false
			}

			stopLoopCounter++
			currentStreamIndex++
			return

		case LF, CR:
			lineCounter, characterCounter = lineCounter+1, 0

		case TRACE:
			command.add(&Command{func([]*Command) {
				cgreader.Tracef("[%d] = %d\n", programIndex, programBuffer[programIndex])
			}, nil})
		}
	}
}

func InitializeParser(input []byte) {
	rawProgramStream, currentStreamIndex = input, 0
	startLoopCounter, stopLoopCounter = 0, 0
	streamIsValid = true
}

func ParseLinearProgram(input []byte) *Command {
	InitializeParser(input)
	command := CreateLinearGroup()
	RecursiveParser(command)
	return command
}

func ParseManualProgram(stream []byte) (*Command, bool) {
	lineCounter, characterCounter = 0, 0
	program := ParseLinearProgram(stream)
	return program, streamIsValid
}

func ParseTargetProgram(stream []byte) (initial, update *Command, result bool) {
	lineCounter, characterCounter = 0, 0
	if index := strings.Index(string(stream), SEPERATOR); index != -1 {
		if initial = ParseLinearProgram(stream[:index-1]); streamIsValid {
			update = ParseLinearProgram(stream[index+3:])
			result = streamIsValid
		} else {
			result = false
		}
	} else {
		fmt.Printf("ERROR! Please seperate your intial and update logic with \"%s\"\n", SEPERATOR)
		result = false
	}
	return
}
