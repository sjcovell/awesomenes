package awesomenes

import (
  "fmt"
  "log"
)

const (
  StatusFlagC = iota
  StatusFlagZ
  StatusFlagI
  StatusFlagD
  StatusFlagB
  StatusFlagU
  StatusFlagV
  StatusFlagN
)

const (
  MemStackBase = 0x10ff
)

type registers struct {
  PC uint16
  SP uint8
  A  uint8
  X  uint8
  Y  uint8
  P  uint8
}

type CPU struct {
  regs *registers
  mem  AddrSpace
}

func MakeCPU(addrSpace AddrSpace) *CPU {
  return &CPU{
    // Top of the stack
    regs:  &registers{
      SP: 0xfd,
    },
    mem: addrSpace,
  }
}

func (cpu *CPU) PowerUp() {
  // Set PC to the reset interrupt vector
  cpu.regs.PC = cpu.mem.Read16(0xfffc)
  cpu.regs.P = 0x24
}

func (cpu *CPU) Run() int {
  fmt.Printf("%v", cpu)

  pcBkp := cpu.regs.PC

  opcode := cpu.mem.Read8(cpu.regs.PC)
  instr, ok := instrTable[opcode]

  if !ok {
    log.Fatalf("Unsupported opcode: %x", opcode)
  }

  instr.fn(cpu, instr.addrMode)

  if cpu.regs.PC == pcBkp {
    cpu.regs.PC += uint16(instr.size)
  }

  return int(instr.cycles)
}

// Same format as the awesome github.com/fogleman/nes for debugging
func (cpu *CPU) String() string {
  opcode := cpu.mem.Read8(cpu.regs.PC)
  instr := instrTable[opcode]
  w0 := fmt.Sprintf("%02X", cpu.mem.Read8(cpu.regs.PC+0))
  w1 := fmt.Sprintf("%02X", cpu.mem.Read8(cpu.regs.PC+1))
  w2 := fmt.Sprintf("%02X", cpu.mem.Read8(cpu.regs.PC+2))
  if instr.size < 2 {
    w1 = "  "
  }
  if instr.size < 3 {
    w2 = "  "
  }
  return fmt.Sprintf(
    "%4X  %s %s %s  %s %28s"+
      "A:%02X X:%02X Y:%02X P:%02X SP:%02X CYC:\n",
    cpu.regs.PC, w0, w1, w2, instr.name, "",
    cpu.regs.A, cpu.regs.X, cpu.regs.Y, cpu.regs.P, cpu.regs.SP)
}

func (c *CPU) Push8(v uint8) {
  c.mem.Write8(c.stackPos(), v)
  c.regs.SP--
}

func (c *CPU) Push16(v uint16) {
  c.Push8(uint8(v >> 8))
  c.Push8(uint8(v & 0xff))
}

func (c *CPU) Pop8() uint8 {
  c.regs.SP++
  return c.mem.Read8(c.stackPos())
}

func (c *CPU) Pop16() uint16 {
  lsb := uint16(c.Pop8())
  msb := uint16(c.Pop8())
  return (msb << 8) + lsb
}

func (c *CPU) stackPos() uint16 {
  return MemStackBase - uint16(c.regs.SP)
}

func (c *CPU) getFlag(flag uint8) bool {
  return (c.regs.P & (0x1 << flag)) != 0
}

func (c *CPU) setFlag(flag uint8) {
  c.regs.P |= (0x1 << flag)
}

func (c *CPU) resetFlag(flag uint8) {
  c.regs.P &= ^(0x1 << flag)
}

func (c *CPU) setOrReset(flag uint8, cond bool) {
  if cond {
    c.setFlag(flag)
  } else {
    c.resetFlag(flag)
  }
}
