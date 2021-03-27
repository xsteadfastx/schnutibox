// https://medium.com/coinmonks/iot-tutorial-read-tags-from-a-usb-rfid-reader-with-raspberry-pi-and-node-red-from-scratch-4554836be127
// nolint: gochecknoglobals, lll, forbidigo, godox
package main

import (
	"fmt"
	"log"

	"github.com/karalabe/hid"
)

const (
	product = 0x27db
	vendor  = 0x16c0
)

const newLine = 40

var charMap = map[byte]string{
	30: "1",
	31: "2",
	32: "3",
	33: "4",
	34: "5",
	35: "6",
	36: "7",
	37: "8",
	38: "9",
	39: "0",
}

func main() {
	readerInfo := hid.Enumerate(vendor, product)[0]

	d, err := readerInfo.Open()
	if err != nil {
		log.Fatal(err)
	}

	id := make(chan string)

	go func() {
		buf := make([]byte, 3)

		rfid := ""

		for {
			// Reading the RFID reader.
			// TODO: Check if "/dev/hidraw" can be used.
			_, err := d.Read(buf)
			if err != nil {
				log.Print(err)

				continue
			}

			// 40 means "\n" and is an indicator that the input is complete.
			// Send the id string to the consumer channel.
			if buf[2] == newLine {
				id <- rfid
				rfid = ""

				continue
			}

			// Check if byte is in the charMap and if its true, append it to the rfid string.
			// It waits till next read and checks for the newline byte.
			c, ok := charMap[buf[2]]
			if ok {
				rfid += c
			}
		}
	}()

	// Wait for an RFID ID and do something with it.
	for {
		rfid := <-id
		fmt.Println(rfid)
	}
}
