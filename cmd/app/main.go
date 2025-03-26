package main

import (
	"log"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
	"splc780d1"
	"time"
)

const i2cBus = "1"
const displayI2cAddress = 0x38

func main() {
	err := initHost()
	if err != nil {
		log.Fatal(err)
	}

	bus, err := newBus(i2cBus)
	if err != nil {
		log.Fatal(err)
	}
	defer bus.Close()

	dev, err := hd4470_i2c.New(bus, displayI2cAddress)
	if err != nil {
		panic(err)
	}

	err = dev.Print("ABC")
	if err != nil {
		panic(err)
	}

	time.Sleep(5 * time.Second)
	err = dev.Reset()
	if err != nil {
		panic(err)
	}

	err = dev.Print("XYZ")
	if err != nil {
		panic(err)
	}
}

func initHost() error {
	if _, err := host.Init(); err != nil {
		return err
	}

	return nil
}

func newBus(i2cBus string) (i2c.BusCloser, error) {
	bus, err := i2creg.Open(i2cBus)
	if err != nil {
		return nil, err
	}

	return bus, err
}
