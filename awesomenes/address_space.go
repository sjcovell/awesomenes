package awesomenes

import (
  "log"
)

type AddrSpace interface {
  Read8(addr uint16) uint8
  Write8(addr uint16, v uint8)

  Read16(addr uint16) uint16
  Read16Bug(addr uint16) uint16
  Write16(addr uint16, v uint16)
}

type CPUAddrSpace struct {
  RAM Memory
  ROM *Rom

  PPU *PPU

  //APU

  // Logger for tests
  Logger Memory

  // Mapper
  // http://tuxnes.sourceforge.net/nesmapper.txt
}

func MakeCPUAddrSpace(rom *Rom, ppu *PPU) *CPUAddrSpace {
  return &CPUAddrSpace{
    RAM:    make(Memory, 0x800),
    ROM:    rom,
    PPU:    ppu,
    Logger: make(Memory, 0x1000),
  }
}

//http://wiki.nesdev.com/w/index.php/CPU_memory_map
//https://wiki.nesdev.com/w/index.php/NROM (Hard coded mapper 0 for now)
func (as *CPUAddrSpace) Read8(addr uint16) uint8 {
  //log.Printf("Reading CPU space %x", addr)
  switch {
    case addr >= 0 && addr < 0x2000:
      // 0x0800 - 0x1fff mirrors 0x0000 - 0x07ff three times
      return as.RAM.Read8(addr % 0x800)

    // PPU registers
    case addr >= 0x2000 && addr < 0x4000:
      //log.Printf("Reading PPUADDR %x", 0x2000 + addr % 8)
      switch 0x2000 + addr % 8 {
        case 0x2002:
          as.PPU.ADDR.SetOnSTATUSRead()
          return as.PPU.STATUS.Get()

        case 0x2004:
          return as.PPU.ReadOAMData()

        case 0x2007:
          return as.PPU.ReadData()

        default:
          log.Fatalf("Invalid read from CPU mem space at %x", addr)
          return 0
      }

    case addr == 0x4015:
        log.Printf("Not yet handled read to APU at %x", addr)
        return 0

    case addr == 0x4016:
        log.Printf("Not yet handled read to controller #1 at %x", addr)
        return 0

    case addr == 0x4017:
        log.Printf("Not yet handled read to controller #2 at %x", addr)
        return 0

    // PRGRAM mirrorred every 0x800 bytes
    case addr >= 0x6000 && addr < 0x8000:
      return as.ROM.PRGRAM.Read8((addr - 0x6000) % 0x800)
      //return as.ROM.PRGRAM.Read8(addr - 0x6000)

    // ROM PRG banks
    case addr >= 0x8000:
      // SRAM mirrorred every 0x800 bytes
      return as.ROM.PRGROM.Read8((addr - 0x8000) % 0x4000)

    default:
      log.Fatalf("Invalid read from CPU mem space at %x", addr)
      return 0
  }
}

func (as *CPUAddrSpace) Write8(addr uint16, v uint8) {

  switch {
    case addr >= 0 && addr < 0x2000:
      if addr == 0x02 || addr == 0x03 {
        log.Printf("WROTE LOG %x at %x", v, addr)
      }
      // 0x0800 - 0x1fff mirrors 0x0000 - 0x07ff three times
      as.RAM.Write8(addr % 0x800, v)

    // PPU registers
    case addr >= 0x2000 && addr < 0x4000:
      as.PPU.STATUS.LastWrite = v

      //log.Printf("Writing PPUADDR %x: %b", addr, v)
      switch 0x2000 + addr % 8 {
        case 0x2000:
          as.PPU.CTRL.Set(v)
          as.PPU.ADDR.SetOnCTRLWrite(v)

        case 0x2001:
          as.PPU.MASK.Set(v)

        case 0x2003:
          as.PPU.OAMADDR = v

        case 0x2004:
          as.PPU.WriteOAMData(v)

        case 0x2005:
          as.PPU.ADDR.SetOnSCROLLWrite(v)
          as.PPU.SCRL.Write(v)

        case 0x2006:
          as.PPU.ADDR.Write(v)

        case 0x2007:
          as.PPU.WriteData(v)

        default:
          log.Fatalf("Invalid write to CPU mem space at %x", addr)
      }

    case addr >= 0x4000 && addr <= 0x4013:
      log.Printf("Not yet handled write to APU at %x", addr)

    case addr == 0x4014:
      //Might need change with mapper
      data  := make([]uint8, 256)
      for i := range(data) {
        data[i] = as.Read8(uint16(v) << 8 + uint16(i))
      }
      as.PPU.OMADMA(data)

    case addr == 0x4015:
      log.Printf("Not yet handled write to APU at %x", addr)

    case addr == 0x4016:
      log.Printf("Not yet handled write to controllers at %x", addr)

    case addr == 0x4017:
      log.Printf("Not yet handled write to APU at %x", addr)

    // PRGRAM mirrorred every 0x800 bytes
    // No CHR RAM for now
    case addr >= 0x6000 && addr < 0x8000:
      log.Printf("Writing to PRGRAM at %x", addr)
      as.ROM.PRGRAM.Write8((addr - 0x6000) % 0x800, v)
      //as.ROM.PRGRAM.Write8(addr - 0x6000, v)

    default:
      //log.Fatalf("Invalid write to CPU mem space at %x", addr)
      log.Printf("Invalid write to CPU mem space at %x", addr)
  }
}

// Little-endian mem layout
func (as *CPUAddrSpace) Read16(addr uint16) uint16 {
  //log.Printf("Reading from %x", addr)
  lo := uint16(as.Read8(addr))
  hi := uint16(as.Read8(addr + 1))
  return (hi << 8) + lo
}

func (as *CPUAddrSpace) Read16Bug(addr uint16) uint16 {
  a := addr
  b := (a & 0xFF00) | uint16(byte(a)+1)
  lo := as.Read8(a)
  hi := as.Read8(b)
  return uint16(hi)<<8 | uint16(lo)
}

func (as *CPUAddrSpace) Write16(addr uint16, v uint16) {
  as.Write8(addr, uint8(v & 0xff))
  as.Write8(addr + 1, uint8(v >> 8))
}
