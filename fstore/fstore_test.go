package fstore

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
)

func NewFirestoreTestClient(ctx context.Context) *FirestoreClient {
	err := os.Setenv("PROJECT", "moda-viewer")
	if err != nil {
		log.Fatal(err)
	}
	err = os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8080")
	if err != nil {
		log.Fatal(err)
	}

	client, err := firestore.NewClient(ctx, "moda-viewer")
	if err != nil {
		log.Fatal(err)
	}
	return &FirestoreClient{Client: client}
}

/* func TestMain(m *testing.M) {
	// command to start firestore emulator
	cmd := exec.Command("firebase", "emulators:start")

	// this makes it killable
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// we need to capture it's output to know when it's started
	stderr, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	defer stderr.Close()

	log.Println("CMD START")
	// start her up!
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	log.Println("DEFER EXIT")
	// ensure the process is killed when we're finished, even if an error occurs
	// (thanks to Brian Moran for suggestion)
	var result int
	defer func() {
		log.Println("DEFER KILL RUN")
		syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		os.Exit(result)
	}()

	// we're going to wait until it's running to start
	var wg sync.WaitGroup
	wg.Add(1)

	log.Println("SEPARATE GOROUTINE")
	// by starting a separate go routine
	go func() {
		// reading it's output
		buf := make([]byte, 256)
		for {
			log.Println("READ START ")
			n, err := stderr.Read(buf[:])
			if err != nil {
				// until it ends
				if err == io.EOF {
					break
				}
				log.Fatalf("reading stderr %v", err)
			}

			log.Println(" IF N")
			if n > 0 {
				d := string(buf[:n])

				// only required if we want to see the emulator output
				log.Printf("%s", d)

				// checking for the message that it's started
				if strings.Contains(d, "All emulators ready!") {
					wg.Done()
				}

				// and capturing the FIRESTORE_EMULATOR_HOST value to set
				pos := strings.Index(d, FirestoreEmulatorHost+"=")
				if pos > 0 {
					host := d[pos+len(FirestoreEmulatorHost)+1 : len(d)-1]
					os.Setenv(FirestoreEmulatorHost, host)
				}
			}
		}
	}()

	log.Println(" WAIT")
	// wait until the running message has been received
	wg.Wait()

	// now it's running, we can run our unit tests
	log.Printf("EMULATOR SUCCESS, RUN UNIT TEST")
	result = m.Run()
} */

const FirestoreEmulatorHost = "FIRESTORE_EMULATOR_HOST"
