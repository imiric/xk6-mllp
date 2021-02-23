package mllp

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/loadimpact/k6/lib"
	"github.com/loadimpact/k6/stats"
)

type MLLP struct {
	opts Options
}

type Options struct {
	Host string
	Port int
}

func NewClient(opts *Options) *MLLP {
	return &MLLP{
		opts: Options{
			Host: opts.Host,
			Port: opts.Port,
		}}
}

// Set the given key with the given value and expiration time.
func (m *MLLP) Send(ctx context.Context, file string) error {
	err := m.sendFile(ctx, file)
	if err != nil {
		return err
	}
	return nil
}

const (
	mllpStart = 0x0b
	mllpEnd   = 0x1c
	mllpEnd2  = 0x0d
)

//Send sends a file over MLLP
func (m *MLLP) sendFile(ctx context.Context, file string) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", m.opts.Host, m.opts.Port))
	if err != nil {
		return err
	}
	defer conn.Close()
	fileContents := m.readFile(file)

	// write the actual message
	conn.Write([]byte{mllpStart})
	fmt.Fprintf(conn, fileContents)
	conn.Write([]byte{mllpEnd})
	conn.Write([]byte{mllpEnd2})

	// read response
	reply := make([]byte, 1024)
	_, err = conn.Read(reply)
	if err != nil {
		return err
	}

	state := lib.GetState(ctx)
	err = errors.New("State is nil")

	if state == nil {
		return err
	}
	stats.PushIfNotDone(ctx, state.Samples, stats.Sample{
		Metric: WriterWrites,
		Time:   time.Time{},
		Value:  float64(len(fileContents)),
	})

	return nil
}

func (m *MLLP) readFile(file string) string {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("Reading file failed:", err.Error())
		os.Exit(1)
	}
	return string(content)
}
