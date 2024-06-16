package main

import (
	"fmt"
	"log"
	"time"

	"github.com/eiannone/keyboard"
	"go.bug.st/serial"
)

const ()

var (
	cntr controller = controller{
		controlByte1:       128,
		controlByte2:       128,
		controlTurn:        128,
		controlAccelerator: 128,
	}
)

type controller struct {
	controlByte1       int  // left/right
	controlByte2       int  // forward/backward
	controlAccelerator int  // accelerating up/down .. 90 .. 150 ..
	controlTurn        int  // trurning
	isFastFly          bool // up / arrow up / tacking off
	isFastDrop         bool // down / arrow down / landing
	isEmergencyStop    bool // stop / word stop / off the thrusters
	isCircleTurnEnd    bool // roll / 360 with arrows in cycle /
	isNoHeadMode       bool // vientiane model / dron with arrows icon / ? no one knwon what is it. i see only affect on lights
	isFastReturn       bool //  - / - / did not find usage in origin
	isUnLock           bool //  - / - / did not find usage in origin
	isGyroCorrection   bool // correct / gear icon / ? correction on wind?
}

func (c *controller) access(i int) int {
	return (((c.controlByte1 ^ c.controlByte2) ^ c.controlAccelerator) ^ c.controlTurn) ^ (i & 255)
}

func (c controller) String() string {
	return fmt.Sprintf(`
======
QE acceleration %v .. 90 .. 150 .. 
AD left/right %v
WS forward/backward %v  
ZC rotation %v
=======
1 isFastFly %v up / arrow up / tacking off
2 isFastDrop %v down / arrow down / landing
3 isEmergencyStop %v stop / word stop / off the thrusters
4 isGyroCorrection %v correct / gear icon / ? correction on wind?
=======`,
		c.controlAccelerator, c.controlByte1, c.controlByte2, c.controlTurn,
		c.isFastFly, c.isFastDrop, c.isEmergencyStop, c.isGyroCorrection)
}

func main() {

	if err := keyboard.Open(); err != nil {
		panic(err)
	}
	defer func() {
		_ = keyboard.Close()
	}()

	conn := connection()
	defer conn.Close()

	ticker := time.NewTicker(500 * time.Millisecond)
	go processingPosition(ticker, conn)

	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			panic(err)
		}

		switch key {
		case keyboard.KeyEsc:
			fmt.Println("Выход")
			ticker.Stop()
			return
		case keyboard.KeySpace:
			cntr.controlByte1 = 128
			cntr.controlByte2 = 128
			cntr.controlTurn = 128
			cntr.controlAccelerator = 128
			cntr.isFastFly = false
			cntr.isFastDrop = false
			cntr.isEmergencyStop = false
			cntr.isCircleTurnEnd = false
			cntr.isNoHeadMode = false
			cntr.isFastReturn = false
			cntr.isUnLock = false
			cntr.isGyroCorrection = false
		}

		switch char {
		case 'a':
			cntr.controlByte1 -= 1
		case 'd':
			cntr.controlByte1 += 1

		case 'w':
			cntr.controlByte2 += 1
		case 's':
			cntr.controlByte2 -= 1

		case 'q': //up
			cntr.controlAccelerator += 1
		case 'e': //down
			cntr.controlAccelerator -= 1

		case 'z':
			cntr.controlTurn += 1
		case 'c':
			cntr.controlTurn -= 1

		case '1':
			cntr.isFastFly = !cntr.isFastFly
		case '2':
			cntr.isFastDrop = !cntr.isFastDrop
		case '3':
			cntr.isEmergencyStop = !cntr.isEmergencyStop
		case '4':
			cntr.isGyroCorrection = !cntr.isGyroCorrection
			// case 't':
			// 	cntr.isCircleTurnEnd = !cntr.isCircleTurnEnd
			// case 'y':
			// 	cntr.isNoHeadMode = !cntr.isNoHeadMode
			// case 'u':
			// 	cntr.isFastReturn = !cntr.isFastReturn
			// case 'i':
			// 	cntr.isUnLock = !cntr.isUnLock
		}
	}
}

func processingPosition(ticker *time.Ticker, conn serial.Port) {
	var snd_cnt int = 0
	for {
		select {
		case <-ticker.C:

			i := 0
			if cntr.isFastFly {
				i = 1
			}
			if cntr.isFastDrop {
				i += 2
			}
			if cntr.isEmergencyStop {
				i += 4
			}
			if cntr.isCircleTurnEnd {
				i += 8
			}
			if cntr.isNoHeadMode {
				i += 16
			}
			if cntr.isFastReturn || cntr.isUnLock {
				i += 32
			}
			if cntr.isGyroCorrection {
				i += 128
			}

			if cntr.controlTurn >= 104 && cntr.controlTurn <= 152 {
				// cntr.controlTurn = 128
			} else if cntr.controlTurn > 255 {
				cntr.controlTurn = 255
			} else if cntr.controlTurn < 1 {
				cntr.controlTurn = 1
			}
			if cntr.controlAccelerator == 1 {
				cntr.controlAccelerator = 0
			}
			if cntr.controlByte1 > 255 {
				cntr.controlByte1 = 255
			} else if cntr.controlByte1 < 1 {
				cntr.controlByte1 = 1
			}
			if cntr.controlByte2 > 255 {
				cntr.controlByte2 = 255
			} else if cntr.controlByte2 < 1 {
				cntr.controlByte2 = 1
			}
			j := i

			fmt.Printf("\n\n\ntick %v\n", cntr)
			snd_cnt = snd_cnt + 1
			if snd_cnt > 32 {
				j = 2
				snd_cnt = 0
			}
			fmt.Printf("snd_cnt = %v", snd_cnt)
			if true {
				send(conn, []byte{
					102,
				})
				send(conn, []byte{
					byte(cntr.controlByte1),
					byte(cntr.controlByte2),
					byte(cntr.controlAccelerator),
					byte(cntr.controlTurn),
					byte(j),
					byte(cntr.access(i)),
					byte(153),
				})
			}
		}
	}
}

func connection() serial.Port {
	// Retrieve the port list
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		log.Fatal("No serial ports found!")
	}

	// Print the list of detected ports
	for _, port := range ports {
		fmt.Printf("Found port: %v\n", port)
	}

	// Open the first serial port detected at 9600bps N81
	mode := &serial.Mode{
		// BaudRate: 9600,
		BaudRate: 19200, //?
		// BaudRate: 57600,
		// BaudRate: 115200,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(ports[0], mode)
	if err != nil {
		log.Fatal(err)
	}

	return port
}

func send(conn serial.Port, bts []byte) {
	fmt.Printf("\n-> %v", bts)

	var n int
	var err error

	if false {
		n, err = conn.Write(bts)
	}

	fmt.Printf("\tn= %v", n)

	if err != nil {
		panic(fmt.Sprint("err:", err))
	}
}
