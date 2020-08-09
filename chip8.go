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

func (vm *chip8) fetchOpcode() uint16 {

	opHigh := uint16(vm.memory[vm.pc]) << 8
	opLow := uint16(vm.memory[vm.pc+1])
	vm.opcode = opHigh | opLow
	vm.pc += 2
	return vm.opcode

}

func (vm *chip8) step() {

	instruction := vm.fetchOpcode()
	completeByte := byte(instruction & 0xFF)
	nibble := byte(instruction & 0xF)
	address := instruction & 0xFFF
	x := byte(instruction >> 8 & 0xF) // check this later maybe will give me some errors
	y := byte(instruction >> 4 & 0xF)

	// only used for debuggin purpose
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
		vm.showSprite(x, y, nibble)
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
	fmt.Println("Clear screen")
}

func (vm *chip8) ret() {
	fmt.Println("Return from subroutine")
}

func (vm *chip8) jump(address uint16) {
	fmt.Println("Jump function")
}

func (vm *chip8) call(address uint16) {
	fmt.Println("call subroutine")
}

func (vm *chip8) skipIfX(vx byte, kk byte) {
	fmt.Println("SkipifX")
}

func (vm *chip8) skipIfNotX(vx byte, kk byte) {
	fmt.Println("SkipIfNot X")
}

func (vm *chip8) skipIfXY(vx byte, vy byte) {
	fmt.Println("SkipIfXY")
}

func (vm *chip8) loadValueInX(vx byte, kk byte) {
	fmt.Println("Carga un valor en el registro x")
}

func (vm *chip8) addValueToX(vx byte, kk byte) {
	fmt.Println("Add kk value to x register and result put in x")
}

func (vm *chip8) copyYtoX(vx byte, vy byte) {
	fmt.Println("copy vy value to vx register")
}

func (vm *chip8) or(vx byte, vy byte) {
	fmt.Println("or operation between vx and vy")
}

func (vm *chip8) and(vx byte, vy byte) {
	fmt.Println("and operation between vx and vy")
}

func (vm *chip8) xor(vx byte, vy byte) {
	fmt.Println("xor between vx and vy")
}

func (vm *chip8) add(vx byte, vy byte) {
	fmt.Println("add between vx and vy")
}

func (vm *chip8) subXY(vx byte, vy byte) {
	fmt.Println("substract vy from vx")
}

func (vm *chip8) shiftRight(vx byte) {
	fmt.Println("Shift vx element to the right (divide by 2) ")
}

func (vm *chip8) subYX(vx byte, vy byte) {
	fmt.Println("substract vx from vy")
}

func (vm *chip8) shiftLeft(vx byte) {
	fmt.Println("Shift vx element to the right (divide by 2) ")
}

func (vm *chip8) compareXY(vx byte, vy byte) {
	fmt.Println("compare x y are not equals")
}

func (vm *chip8) setI(address uint16) {
	fmt.Println("set I register with vx value")
}

func (vm *chip8) jumpTo(address uint16) {
	fmt.Println("jump to specified address")
}

func (vm *chip8) random(vx byte, kk byte) {
	fmt.Println("create a random number and put it in vx")
}

func (vm *chip8) showSprite(vx byte, vy byte, nibble byte) {
	fmt.Println("display n byte sprite starting at i memory")
}

func (vm *chip8) skipIfPressed(vx byte) {
	fmt.Println("skip instruction if vx value is equal to keyboard pressed")
}

func (vm *chip8) skipIfNotPressed(vx byte) {
	fmt.Println("skip instruction if vx value is not equal to keyboard pressed")
}

func (vm *chip8) putTimerInX(vx byte) {
	fmt.Println("put value from dst register in vx")
}

func (vm *chip8) waitForKeyPress(vx byte) {
	fmt.Println("Wait for key press, store key value in vx")
}

func (vm *chip8) setDelay(vx byte) {
	fmt.Println("Dt is set to vx value")
}

func (vm *chip8) setSound(vx byte) {
	fmt.Println("st is set to vx value")
}

func (vm *chip8) addXToI(vx byte) {
	fmt.Println("vx and i are added results stored in I")
}

func (vm *chip8) loadF(vx byte) {
	fmt.Println("i is set to the location of hexadecimal representation ofthe vx value")
}

func (vm *chip8) loadBCD(vx byte) {
	fmt.Println("Store representation of hexadecimal vx in I")
}

func (vm *chip8) saveRegisters(vx byte) {
	fmt.Println("store al v0 .... vx register in memory starting at I location")
}

func (vm *chip8) loadRegisters(vx byte) {
	fmt.Println("read value from memory starting at I into register v0 through vx")
}

func (vm *chip8) writeToMem(high byte, low byte) {
	vm.memory[512] = high
	vm.memory[513] = low
}

func main() {
	chip := chip8{}
	chip.Initialize()
	chip.writeToMem(0x32, 0x20)
	fmt.Println(chip)
	chip.step()
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
