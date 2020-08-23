package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
)

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
	sp         uint          // stack pointer, point to last expression inside stack
	i          uint16        // to store memory address
	pc         uint16        // show the addres where actual program is
	stack      [16]uint16    // stack, support 16 nested stack calls
	display    [64 * 32]byte // display mapper
	dt         byte          // delay timer
	st         byte          // sound timer
	keyboard   [16]byte      // keyboard keys goes from 0 to F
	drawScreen bool          // flag to set when to clear and draw in screen
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
	fmt.Println("Loading ROM......")
	vm.loadROM("text.txt")

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

func (vm *chip8) fetchOpcode() uint16 {

	opHigh := uint16(vm.memory[vm.pc]) << 8
	opLow := uint16(vm.memory[vm.pc+1])
	vm.opcode = opHigh | opLow
	vm.pc += 2
	// To ensure not get a overflow for dummy logic
	// remove later
	if vm.pc > 4095 {
		vm.pc = 0x200
	}
	return vm.opcode

}

func (vm *chip8) step() {

	instruction := vm.fetchOpcode()
	completeByte := byte(instruction & 0xFF)
	nibble := byte(instruction & 0xF)
	address := instruction & 0xFFF
	x := uint(instruction >> 8 & 0xF) // check this, later maybe will give me some errors
	y := uint(instruction >> 4 & 0xF)

	// only used for debugging purpose
	debug(instruction)

	switch {
	case instruction == 0x00E0:
		vm.cls()
	case instruction == 0x00EE:
		vm.ret()
	case instruction&0xF000 == 0x1000:
		vm.jump(address)
	case instruction&0xF000 == 0x2000:
		vm.call(address)
	case instruction&0xF000 == 0x3000:
		vm.skipIfX(x, completeByte)
	case instruction&0xF000 == 0x4000:
		vm.skipIfNotX(x, completeByte)
	case instruction&0xF000 == 0x5000:
		vm.skipIfXY(x, y)
	case instruction&0xF000 == 0x6000:
		vm.loadValueInX(x, completeByte)
	case instruction&0xF000 == 0x7000:
		vm.addValueToX(x, completeByte)
	case instruction&0xF000 == 0x8000:
		vm.copyYtoX(x, y)
	case instruction&0XF00F == 0x8001:
		vm.or(x, y)
	case instruction&0xF00F == 0x8002:
		vm.and(x, y)
	case instruction&0xF00F == 0x8003:
		vm.xor(x, y)
	case instruction&0xF00F == 0x8004:
		vm.add(x, y)
	case instruction&0xF00F == 0x8005:
		vm.subXY(x, y)
	case instruction&0xF00F == 0x8006:
		vm.shiftRight(x)
	case instruction&0xF00F == 0x8007:
		vm.subYX(x, y)
	case instruction&0xF00F == 0x800E:
		vm.shiftLeft(x)
	case instruction&0xF00F == 0x9000:
		vm.compareXY(x, y)
	case instruction&0xF000 == 0xA000:
		vm.setI(address)
	case instruction&0xF000 == 0xB000:
		vm.jumpTo(address)
	case instruction&0xF000 == 0xC000:
		vm.random(x, completeByte)
	case instruction&0xF000 == 0xD000:
		vm.drawSprite(x, y, nibble)
	case instruction&0xF0FF == 0xE09E:
		vm.skipIfPressed(x)
	case instruction&0xF0FF == 0xE0A1:
		vm.skipIfNotPressed(x)
	case instruction&0xF00F == 0xF007:
		vm.putTimerInX(x)
	case instruction&0xF00F == 0xF00A:
		vm.waitForKeyPress(x)
	case instruction&0xF0FF == 0xF015:
		vm.setDelay(x)
	case instruction&0xF0FF == 0xF018:
		vm.setSound(x)
	case instruction&0xF0FF == 0xF01E:
		vm.addXToI(x)
	case instruction&0xF0FF == 0xF029:
		vm.loadF(x)
	case instruction&0xF0FF == 0xF033:
		vm.loadBCD(x)
	case instruction&0xF0FF == 0xF055:
		vm.saveRegisters(x)
	case instruction&0xF0FF == 0xF065:
		vm.loadRegisters(x)

	}

}

// Opcode methods
func (vm *chip8) cls() {
	/*
		Clear the whole display and set draw flag to true
		to update it.
	*/
	vm.clearDisplay()
	vm.drawScreen = true

}

func (vm *chip8) ret() {
	/*
		The interpreter sets the program counter
		to the address at the top of the stack
	*/
	if int(vm.sp) == 0 {
		fmt.Println("Stack underflow")
	}
	vm.sp--
	vm.pc = vm.stack[vm.sp]

}

func (vm *chip8) jump(address uint16) {
	/*
		interpreter sets the program counter to input address
	*/
	vm.pc = address
}

func (vm *chip8) call(address uint16) {
	/*
		Save current pc in stack and then jump to address
		(call a subroutine)
	*/
	if int(vm.sp) > len(vm.stack) {
		fmt.Println("Stack overflow!!!")
	}
	vm.stack[vm.sp] = vm.pc
	vm.sp++
	vm.jump(address)
}

func (vm *chip8) skipIfX(vx uint, kk byte) {
	/*
		The interpreter compares register Vx to kk, and if they are equal,
		increments the program counter by 2.
	*/
	if vm.v[vx] == kk {
		vm.pc += 2
	}
}

func (vm *chip8) skipIfNotX(vx uint, kk byte) {
	/*
		Compares register Vx to kk, and if they are not equal,
		increments the program counter by 2.
	*/
	if vm.v[vx] != kk {
		vm.pc += 2
	}
}

func (vm *chip8) skipIfXY(vx uint, vy uint) {
	/*
		Compares register Vx to register Vy, and if they are equal,
		increments the program counter by 2.
	*/
	if vm.v[vx] == vm.v[vy] {
		vm.pc += 2
	}
}

func (vm *chip8) loadValueInX(vx uint, kk byte) {
	/*
		Interpreter puts the value kk into register Vx.
	*/
	vm.v[vx] = kk
}

func (vm *chip8) addValueToX(vx uint, kk byte) {
	/*
		Adds the value kk to the value of register Vx, then stores the result in Vx.
	*/
	vm.v[vx] += kk
}

func (vm *chip8) copyYtoX(vx uint, vy uint) {
	/*
		Stores the value of register Vy in register Vx.
	*/
	vm.v[vx] = vm.v[vy]
}

func (vm *chip8) or(vx uint, vy uint) {
	/*
		Performs a bitwise OR on the values of Vx and Vy, then stores the result in Vx
	*/
	vm.v[vx] = vm.v[vx] | vm.v[vy]

}

func (vm *chip8) and(vx uint, vy uint) {
	/*
		Performs a bitwise AND on the values of Vx and Vy, then stores the result in Vx.
	*/
	vm.v[vx] = vm.v[vx] & vm.v[vy]

}

func (vm *chip8) xor(vx uint, vy uint) {
	/*
		Performs a bitwise exclusive OR on the values of Vx and Vy, then stores the result in Vx
	*/
	vm.v[vx] = vm.v[vx] ^ vm.v[vy]
}

func (vm *chip8) add(vx uint, vy uint) {
	/*
		The values of Vx and Vy are added together. If the result is greater than 8 bits (i.e., > 255,)
		VF is set to 1, otherwise 0. Only the lowest 8 bits of the result are kept, and stored in Vx.
	*/

	// if a overflow happens go already give you the lowest 8 bits of the operation
	vm.v[vx] += vm.v[vy]

	// the only way in which v[x] + v[y] is less than v[y] is in and overflow
	// when v[x] could go from 255 to 0, 1, 3, etc
	if vm.v[vx] < vm.v[vy] {
		vm.v[0xF] = 1
	} else {
		vm.v[0xF] = 0
	}

}

func (vm *chip8) subXY(vx uint, vy uint) {
	/*
		If Vx > Vy, then VF is set to 1, otherwise 0.
		Then Vy is subtracted from Vx, and the results stored in Vx.
	*/
	if vm.v[vx] > vm.v[vy] {
		vm.v[0xF] = 1
	} else {
		vm.v[0xF] = 0
	}
	vm.v[vx] -= vm.v[vy]

}

func (vm *chip8) shiftRight(vx uint) {
	/*
		If the least-significant bit of Vx is 1,
		then VF is set to 1, otherwise 0. Then Vx is divided by 2.
	*/
	vm.v[0xF] = vm.v[vx] & 0x01 // if we do a masking here is not needed to do some if to test equality
	vm.v[vx] >>= 1
}

func (vm *chip8) subYX(vx uint, vy uint) {
	/*
		If Vy > Vx, then VF is set to 1, otherwise 0.
		Then Vx is subtracted from Vy, and the results stored in Vx.
	*/
	if vm.v[vy] > vm.v[vx] {
		vm.v[0x0F] = 1
	} else {
		vm.v[0x0F] = 0
	}
	vm.v[vx] = vm.v[vy] - vm.v[vx]
}

func (vm *chip8) shiftLeft(vx uint) {
	/*
		If the most-significant bit of Vx is 1, then VF is set to 1, otherwise to 0.
		Then Vx is multiplied by 2.
	*/
	vm.v[0xF] = vm.v[vx] >> 7
	vm.v[vx] <<= 1
}

func (vm *chip8) compareXY(vx uint, vy uint) {
	/*
		The values of Vx and Vy are compared, and if they are not equal,
		the program counter is increased by 2
	*/
	if vm.v[vx] != vm.v[vy] {
		vm.pc += 2
	}

}

func (vm *chip8) setI(address uint16) {
	/*
		The value of register I is set to address
	*/
	vm.i = address
}

func (vm *chip8) jumpTo(address uint16) {
	/*
		The program counter is set to nnn plus the value of V0
	*/
	vm.pc = vm.pc + uint16(vm.v[0])
}

func (vm *chip8) random(vx uint, kk byte) {
	/*

		The interpreter generates a random number from 0 to 255, which is then ANDed with the value kk.
		The results are stored in Vx
	*/
	random := byte(rand.Intn(255))
	vm.v[vx] = random & kk
}

func (vm *chip8) drawSprite(vx uint, vy uint, nibble byte) {
	/*
		The interpreter reads n bytes from memory, starting at the address stored in I.
		These bytes are then displayed as sprites on screen at coordinates (Vx, Vy).
		Sprites are XORed onto the existing screen. If this causes any pixels to be erased,
		VF is set to 1, otherwise it is set to 0. If the sprite is positioned so part of it is outside the
		coordinates of the display, it wraps around to the opposite side of the screen.
	*/

	x := uint16(vm.v[vx])
	y := uint16(vm.v[vy])
	n := uint16(nibble)

	var xline uint16
	var yline uint16
	var pixel byte

	for yline = 0; yline < n; y++ {
		pixel = vm.memory[vm.i+yline]
		for xline = 0; xline < 8; xline++ { // xline lenght is fixed
			if pixel&(0x80>>xline) != 0 {
				if (x + xline + ((y + yline) * 64)) < 2048 { // store in column-major order
					if vm.display[(x+xline+((y+yline)*64))] == 1 {
						vm.v[0xF] = 1
					}
					vm.display[x+xline+((y+yline)*64)] ^= 0x1
				}
			}
		}
	}
	vm.drawScreen = true
}

func (vm *chip8) skipIfPressed(vx uint) {
	/*
		Skip next instruction if key with the value of Vx is pressed.
		Checks the keyboard, and if the key corresponding to the value of Vx is currently in the down position
		PC is increased by 2.
	*/
	if vm.keyboard[vm.v[vx]] != 0 {
		vm.pc += 2
	}
}

func (vm *chip8) skipIfNotPressed(vx uint) {
	/*
		Skip next instruction if key with the value of Vx is not pressed.
		Checks the keyboard, and if the key corresponding to the value of Vx is currently in the up position,
		PC is increased by 2.
	*/
	if vm.keyboard[vm.v[vx]] == 0 {
		vm.pc += 2
	}
}

func (vm *chip8) putTimerInX(vx uint) {

	/*
		The value of DT is placed into Vx.
	*/
	vm.v[vx] = vm.dt
}

func (vm *chip8) waitForKeyPress(vx uint) {
	/*
		All execution stops until a key is pressed, then the value of that key is stored in Vx
	*/

	// need to check here how t ohandle the keyboard
	keyPressed := false
	for i := 0; i < 16; i++ {
		if vm.keyboard[i] != 0 {
			vm.v[vx] = byte(i)
			keyPressed = true
		}
	}
	if keyPressed {
		vm.pc += 2
	}

}

func (vm *chip8) setDelay(vx uint) {
	/*
		DT is set equal to the value of Vx.
	*/
	vm.dt = vm.v[vx]

}

func (vm *chip8) setSound(vx uint) {
	/*
		ST is set equal to the value of Vx.
	*/
	vm.st = vm.v[vx]
}

func (vm *chip8) addXToI(vx uint) {
	/*
		The values of I and Vx are added, and the results are stored in I.
	*/
	vm.i = vm.i + uint16(vm.v[vx])
}

func (vm *chip8) loadF(vx uint) {
	/*
		The value of I is set to the location for the hexadecimal
		sprite corresponding to the value of Vx
	*/
	// Hexadecimal font are loaded in init method to memory (0-80 array range)
	// if we want to set the sprite 2 font address, direction we will have
	// 2*5= 10. The 0 sprite occupies from memory[0] to memory[4], sprite 1 occupies from memory[5] to memory[9]
	// so sprite 2 is read from memory[10]
	vm.i = uint16(vm.v[vx]) * 0x5
	fmt.Println("Set sprite")

}

func (vm *chip8) loadBCD(vx uint) {
	/*
		Store BCD representation of Vx in memory locations I, I+1, and I+2.
		The interpreter takes the decimal value of Vx, and places the hundreds digit in memory at location in I,
		the tens digit at location I+1, and the ones digit at location I+2.
		Implementing double dabble algorithm.
	*/
	x := uint(vm.v[vx])
	bcd := uint(0)

	for n := 0; n < 8; n++ {

		if (bcd>>0)&0xF >= 5 {
			bcd += 3
		}
		if (bcd>>4)&0xF >= 5 {
			bcd = (bcd + 3) << 4
		}
		if (bcd>>8)&0xF >= 5 {
			bcd = (bcd + 3) << 8
		}
		// make a bit shift (some tricks to do the share shift register from double dabble algorithmn)
		bcd = (bcd << 1) | (x >> (7 - n) & 0x1)
	}

	vm.memory[vm.i] = byte(bcd>>8) & 0xF
	vm.memory[vm.i+1] = byte(bcd>>4) & 0xF
	vm.memory[vm.i+2] = byte(bcd>>0) & 0xF

}

func (vm *chip8) saveRegisters(vx uint) {
	/*
		The interpreter copies the values of registers V0 through Vx into memory,
		starting at the address in I.
	*/
	x := uint16(vx)
	for i := uint16(0); i <= x; i++ {
		vm.memory[vm.i+i] = vm.v[i]
	}
}

func (vm *chip8) loadRegisters(vx uint) {
	/*
		The interpreter reads values from memory starting at location I into registers V0 through Vx.
	*/
	x := uint16(vx)
	for i := uint16(0); i <= x; i++ {
		vm.v[i] = vm.memory[vm.i+i]
	}
}

func (vm *chip8) loadROM(path string) error {
	rom, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	baseSpace := 0x200 // see chip8 memory map, normally programs begin here
	romLength := len(rom)
	if romLength > 0x1000-baseSpace {
		return errors.New("Too large to fit in memory")
	}
	for i := 0; i < romLength; i++ {
		// load rom after 512 address
		vm.memory[i+baseSpace] = rom[i]
	}
	return nil
}

func (vm *chip8) setKeys() {
	fmt.Println("Reading keys")
}

func (vm *chip8) writeToMem(high byte, low byte) {
	vm.memory[512] = high
	vm.memory[513] = low
}

func main() {
	/*
		To debug chip8 commands
		chip := chip8{}
		chip.Initialize()
		chip.writeToMem(0x32, 0x20)
		fmt.Println(chip)
		chip.step()
	*/

	chip := chip8{}

	setGraphics()
	setInputHandlers()
	chip.Initialize()
	// load chip8 rom
	chip.loadROM("text.txt")
	fmt.Println(chip.memory)
	for {

		chip.step()
		if chip.drawScreen {
			drawGraphics()
		}
		chip.setKeys()
	}

}

// utilities functions
func debug(instruction uint16) {
	completeByte := instruction & 0xFF
	nibble := instruction & 0xF
	address := instruction & 0xFFF
	x := instruction >> 8 & 0xF
	y := instruction >> 4 & 0xF

	fmt.Printf("Instruction: %x\n", instruction)
	fmt.Printf("Intruction byte: %x\n", completeByte)
	fmt.Printf("Nibble: %x\n", nibble)
	fmt.Printf("Address: %x\n", address)
	fmt.Printf("X: %x\n", x)
	fmt.Printf("Y: %x\n", y)
}

func setGraphics() {
	fmt.Println("Set graphic library")
}

func setInputHandlers() {
	fmt.Println("Set input handlers")
}

func drawGraphics() {
	fmt.Println("Drawing inside screen")
}
