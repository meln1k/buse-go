package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/meln1k/buse-go/buse"
)

// This device is an example implementation of an in-memory block device with high latency of reads/writes

type DeviceExample struct {
	dataset []byte
}

func (d *DeviceExample) ReadAt(p []byte, off uint64) error {
	time.Sleep(5 * time.Millisecond)
	copy(p, d.dataset[off:int(off)+len(p)])
	log.Printf("[DeviceExample] READ offset:%d len:%d\n", off, len(p))
	return nil
}

func (d *DeviceExample) WriteAt(p []byte, off uint64) error {
	time.Sleep(5 * time.Millisecond)
	copy(d.dataset[off:], p)
	log.Printf("[DeviceExample] WRITE offset:%d len:%d\n", off, len(p))
	return nil
}

func (d *DeviceExample) Disconnect() {
	log.Println("[DeviceExample] DISCONNECT")
}

func (d *DeviceExample) Flush() error {
	log.Println("[DeviceExample] FLUSH")
	return nil
}

func (d *DeviceExample) Trim(off uint64, length uint32) error {
	log.Printf("[DeviceExample] TRIM offset:%d len:%d\n", off, length)
	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s /dev/nbd0\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		usage()
	}
	size := uint64(1024 * 1024 * 512) // 512M
	deviceExp := &DeviceExample{}
	deviceExp.dataset = make([]byte, size)
	device, err := buse.CreateDevice(args[0], size, deviceExp, 64)
	if err != nil {
		fmt.Printf("Cannot create device: %s\n", err)
		os.Exit(1)
	}
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	go func() {
		if err := device.Connect(); err != nil {
			log.Printf("Buse device stopped with error: %s", err)
		} else {
			log.Println("Buse device stopped gracefully.")
		}
	}()
	<-sig
	// Received SIGTERM, cleanup
	fmt.Println("SIGINT, disconnecting...")
	device.Disconnect()
}
