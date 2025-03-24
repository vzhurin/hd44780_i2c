package splc780d1

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

	err := dev.init()
	if err != nil {
		return nil, err
	}

	return dev, nil
}

func (d *Dev) WriteString(message string, row int, startPosition byte) error {
	var position byte

	switch row {
	case 1:
		position = startPosition
	case 2:
		position = 0x40 + startPosition
	case 3:
		position = 0x14 + startPosition
	case 4:
		position = 0x54 + startPosition
	}

	err := d.writeCommand(0x80+position, 0)
	if err != nil {

	}

	for _, c := range []byte(message) {
		err = d.writeCommand(c, RS)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Dev) Clear() error {
	err := d.writeCommand(cmdClearDisplay, E)
	if err != nil {
		return err
	}

	return d.writeCommand(cmdReturnHome, E)
}

func (d *Dev) init() error {
	initBytes := []byte{0x03, 0x03, 0x03, 0x02}
	for _, b := range initBytes {
		err := d.writeCommand(b, 0)
		if err != nil {
			return err
		}
	}

	setupBytes := []byte{
		cmdFunctionSet | opt2LinesDisplay | optFont5x8Dots | optDataLength4Bit,
		cmdDisplayControl | optEnableDisplay,
		cmdClearDisplay,
		cmdEntryMode | optIncrement,
	}
	for _, b := range setupBytes {
		err := d.writeCommand(b, 0)
		if err != nil {
			return err
		}
	}

	sleepMs(200)

	return nil
}

func (d *Dev) write(b byte) error {
	err := d.bus.Tx(d.address, []byte{b}, nil)
	if err != nil {
		return nil
	}

	sleepUs(100)

	return nil
}

func (d *Dev) strobe(b byte) error {
	if d.backlight {
		b |= backlight
	}
	sleepUs(200)

	err := d.write(b | E)
	if err != nil {
		return err
	}
	sleepUs(30)

	err = d.write(b &^ E)

	return err
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

func (d *Dev) writeCommand(cmd byte, mode byte) error {
	err := d.write4Bits(mode | (cmd & 0xF0))
	if err != nil {
		return err
	}

	return d.write4Bits(mode | ((cmd << 4) & 0xF0))
}

func sleepUs(d uint) {
	time.Sleep(time.Duration(d) * time.Microsecond)
}

func sleepMs(d uint) {
	time.Sleep(time.Duration(d) * time.Millisecond)
}
