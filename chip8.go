package main

import "fmt"

// := can only be used inside a function, outside we need to declare explicitly
var hexadecimalSprites = [80]byte{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F

}

// Here we can use uint8 or byte, i think  byte is more semantic
type chip8 struct {
	memory     [4096]byte    // 4096 or 4KB of memory from 0x0000 to 0x1000
	v          [16]byte      // general purpose registers from v0, v1, ...., vF
	sp         byte          // stack pointer, point to last expression inside stack
	i          uint16        // to store memory address
	pc         uint16        // show the addres where actual program is
	stack      [16]uint16    // stack, support 16 nested stack calls
	display    [64 * 32]byte // display mapper
	dt         byte          // delay timer
	st         byte          // sound timer
	keyboard   [16]byte      // keyboard keys goes from 0 to F
	drawScreen bool          // flag to set when to clear and draw the screen
	opcode     uint16        // current opcode
}

func (vm *chip8) Initialize() {
	fmt.Println("Initializing chip8 emulator....")
	vm.pc = 0x200 // initialize program counter following specs
	vm.i = 0x00
	vm.sp = 0x00
	vm.opcode = 0x00

	// clear all emulated hardware
	vm.clearDisplay()
	vm.clearStack()
	vm.clearRegisters()
	vm.clearMemory()

	// load font_set
	for i := 0; i < 80; i++ {
		vm.memory[i] = hexadecimalSprites[i]
	}

}

func (vm *chip8) clearDisplay() {
	for i := range vm.display {
		vm.display[i] = 0x00
	}
}

func (vm *chip8) clearRegisters() {
	for i := range vm.v {
		vm.v[i] = 0x00
	}
}

func (vm *chip8) clearStack() {
	for i := range vm.stack {
		vm.stack[i] = 0x00
	}
}

func (vm *chip8) clearMemory() {
	for i := range vm.memory {
		vm.memory[i] = 0x00
	}
}

func main() {
	chip := chip8{}
	chip.Initialize()
	fmt.Println(chip)
}
