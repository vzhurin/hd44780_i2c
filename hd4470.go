package hd4470_i2c

import (
	"periph.io/x/conn/v3/i2c"
	"time"
)

const (
	cmdClearDisplay         = 0x01
	cmdReturnHome           = 0x02
	cmdEntryMode            = 0x04
	cmdDisplayControl       = 0x08
	cmdCursorOrDisplayShift = 0x10
	cmdFunctionSet          = 0x20

	// Entry mode
	optIncrement           = 0x02
	optDecrement           = 0x00
	optDisplayShift        = 0x01
	optDisplayWithoutShift = 0x00

	// Display control
	optEnableDisplay  = 0x04
	optDisableDisplay = 0x00
	optEnableCursor   = 0x02
	optDisableCursor  = 0x00
	optEnableBlink    = 0x01
	optDisableBlink   = 0x00

	// Cursor or display shift
	optShiftCursorLeft   = 0x00
	optShiftCursorRight  = 0x04
	optShiftDisplayLeft  = 0x08
	optShiftDisplayRight = 0x0C

	// Function set
	optDataLength8Bit = 0x10
	optDataLength4Bit = 0x00
	opt2LinesDisplay  = 0x08
	opt1LineDisplay   = 0x00
	optFont5x10Dots   = 0x04
	optFont5x8Dots    = 0x00

	backlight = 0x08

	// E — signal to start reading or writing data.
	E = 0b00000100

	// RW - signal to select read or write action. 1: Read, 0: Write.
	RW = 0b00000010

	// RS — register select signal.
	// 1: Data Register (for read and write)
	// 0: Instruction Register (for write),
	// Busy flag - Address Counter (for read).
	RS = 0b00000001
)

var resetSequence = [][2]uint{
	{0x03, 50}, // init 1-st cycle
	{0x03, 10}, // init 2-nd cycle
	{0x03, 10}, // init 3-rd cycle
	{0x02, 10}, // init finish
}

var initSequence = [][2]uint{
	{0x14, 0},    // 4-bit mode, 2 lines, 5x7 chars high
	{0x10, 0},    // disable display
	{0x01, 2000}, // clear screen
	{0x06, 0},    // cursor shift right, no display move
	{0x0c, 0},    // enable display no cursor
	{0x01, 2000}, // clear screen
	{0x02, 2000}, // cursor home
}

//var initSequence = [][2]uint{
//	{cmdFunctionSet | optDataLength4Bit | opt2LinesDisplay | optFont5x8Dots, 0},
//	{cmdDisplayControl | optDisableDisplay, 0},
//	{cmdClearDisplay, 2000},
//	{cmdCursorOrDisplayShift | optShiftCursorRight, 0},
//	{cmdDisplayControl | optEnableDisplay | optDisableCursor, 0},
//	{cmdClearDisplay, 2000},
//	{cmdReturnHome, 2000},
//}

type Dev struct {
	bus       i2c.Bus
	address   uint16
	backlight bool
}

func New(bus i2c.Bus, address uint16) (*Dev, error) {
	dev := &Dev{
		bus:       bus,
		address:   address,
		backlight: false,
	}

	err := dev.Reset()
	if err != nil {
		return nil, err
	}

	return dev, nil
}

func (d *Dev) Reset() error {
	for _, c := range resetSequence {
		err := d.write4Bits(byte(c[0]))
		if err != nil {
			return err
		}
		sleepUs(c[1])
	}

	for _, c := range initSequence {
		err := d.writeInstruction(byte(c[0]), 0)
		if err != nil {
			return err
		}
		sleepUs(c[1])
	}

	return nil
}

func (d *Dev) Print(data string) error {
	for _, v := range []byte(data) {
		err := d.WriteChar(v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Dev) WriteChar(char byte) error {
	err := d.writeInstruction(char, RS)
	if err != nil {
		return err
	}

	sleepUs(10)

	return nil
}

func (d *Dev) write(b byte) error {
	err := d.bus.Tx(d.address, []byte{b}, nil)
	if err != nil {
		return nil
	}

	return nil
}

func (d *Dev) write4Bits(b byte) error {
	if d.backlight {
		b |= backlight
	}

	err := d.write(b)
	if err != nil {
		return err
	}

	return d.strobe(b)
}

func (d *Dev) writeInstruction(cmd byte, mode byte) error {
	err := d.write4Bits((cmd & 0xF0) | mode)
	if err != nil {
		return err
	}

	err = d.write4Bits(((cmd << 4) & 0xF0) | mode)
	if err != nil {
		return err
	}

	sleepUs(50)

	return nil
}

func (d *Dev) strobe(b byte) error {
	if d.backlight {
		b |= backlight
	}

	err := d.write(b | E)
	if err != nil {
		return err
	}

	sleepUs(2)

	err = d.write(b &^ E)

	return err
}

func sleepUs(d uint) {
	time.Sleep(time.Duration(d) * time.Microsecond)
}

func sleepMs(d uint) {
	time.Sleep(time.Duration(d) * time.Millisecond)
}
