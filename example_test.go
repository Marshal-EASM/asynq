// Copyright 2020 Kentaro Hibino. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package asynq_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"testing"
	"time"

	"github.com/Marshal-EASM/asynq"
	"golang.org/x/sys/unix"
)

// func TestName(t *testing.T) {
//
// }
func TestExampleServerRun(t *testing.T) {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: ":6379", DB: 0},
		asynq.Config{
			StrictPriority: true,
			LogLevel:       1,
			Concurrency:    2,
			Queues: map[string]int{
				"test": 5,
			},
		},
	)

	mux := asynq.NewServeMux()
	// mux.HandleFunc(protocol.Domain, handler.HandleDomain)
	// mux.HandleFunc(protocol.WebScan, handler.WebScan)
	// mux.HandleFunc(protocol.VulnerabilityScan, handler.VulnScan)
	// mux.HandleFunc(protocol.PortScan, handler.PortScan)
	// mux.HandleFunc(protocol.PassiveScan, handler.PassiveScan)
	// mux.HandleFunc(protocol.XssScan, handler.XssScan)
	// mux.HandleFunc(protocol.JsDetect, handler.JsScan)
	mux.HandleFunc("test", ABC)
	// Run blocks and waits for os signal to terminate the program.
	if err := srv.Run(mux); err != nil {
		log.Fatal(err)
	}
}

type WebScanArg struct {
	TaskID     string
	QueueID    string
	Domain     []string
	ScreenShot bool
}

func ABC(ctx context.Context, t *asynq.Task) error {
	fmt.Println("start")
	defer fmt.Println("end")
	// var d WebScanArg
	// if err := json.Unmarshal(t.Payload(), &d); err != nil {
	// 	fmt.Println(err)
	// }
	time.Sleep(5 * time.Second)
	return nil
}

func TestName(t *testing.T) {
	QueueClient := asynq.NewClient(asynq.RedisClientOpt{Addr: ":6379", DB: 0})
	for i := 0; i < 10; i++ {
		_, err := QueueClient.Enqueue(asynq.NewTask("test", nil), asynq.Queue("test"), asynq.MaxRetry(1), asynq.Timeout(60*60*24*30*time.Second))
		if err != nil {
			fmt.Println(err)
		}
	}
}

func ExampleServer_Shutdown() {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: ":6379"},
		asynq.Config{Concurrency: 20},
	)

	h := asynq.NewServeMux()
	// ... Register handlers

	if err := srv.Start(h); err != nil {
		log.Fatal(err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, unix.SIGTERM, unix.SIGINT)
	<-sigs // wait for termination signal

	srv.Shutdown()
}

func ExampleServer_Stop() {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: ":6379"},
		asynq.Config{Concurrency: 20},
	)

	h := asynq.NewServeMux()
	// ... Register handlers

	if err := srv.Start(h); err != nil {
		log.Fatal(err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, unix.SIGTERM, unix.SIGINT, unix.SIGTSTP)
	// Handle SIGTERM, SIGINT to exit the program.
	// Handle SIGTSTP to stop processing new tasks.
	for {
		s := <-sigs
		if s == unix.SIGTSTP {
			srv.Stop() // stop processing new tasks
			continue
		}
		break // received SIGTERM or SIGINT signal
	}

	srv.Shutdown()
}

func ExampleScheduler() {
	scheduler := asynq.NewScheduler(
		asynq.RedisClientOpt{Addr: ":6379"},
		&asynq.SchedulerOpts{Location: time.Local},
	)

	if _, err := scheduler.Register("* * * * *", asynq.NewTask("task1", nil)); err != nil {
		log.Fatal(err)
	}
	if _, err := scheduler.Register("@every 30s", asynq.NewTask("task2", nil)); err != nil {
		log.Fatal(err)
	}

	// Run blocks and waits for os signal to terminate the program.
	if err := scheduler.Run(); err != nil {
		log.Fatal(err)
	}
}

func ExampleParseRedisURI() {
	rconn, err := asynq.ParseRedisURI("redis://localhost:6379/10")
	if err != nil {
		log.Fatal(err)
	}
	r, ok := rconn.(asynq.RedisClientOpt)
	if !ok {
		log.Fatal("unexpected type")
	}
	fmt.Println(r.Addr)
	fmt.Println(r.DB)
	// Output:
	// localhost:6379
	// 10
}

func ExampleResultWriter() {
	// ResultWriter is only accessible in Handler.
	h := func(ctx context.Context, task *asynq.Task) error {
		// .. do task processing work

		res := []byte("task result data")
		n, err := task.ResultWriter().Write(res) // implements io.Writer
		if err != nil {
			return fmt.Errorf("failed to write task result: %v", err)
		}
		log.Printf(" %d bytes written", n)
		return nil
	}

	_ = h
}
