package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"syscall"
	"time"

	"google.golang.org/protobuf/proto"
)

type Res struct {
	Cpu     int64
	Mem     uint64
	Arrived int64
}

func main() {
	http.HandleFunc("/ping", handler)

	fmt.Println("Listening at :1111...")
	http.ListenAndServe(":1111", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	arrive := time.Now()
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)
	cpuStart := CPUTime()

	q := r.URL.Query()
	payload := q.Get("payload")

	b := ReqBody{}
	ser, _ := io.ReadAll(r.Body)
	if payload == "json" {
		json.Unmarshal(ser, &b)
	} else if payload == "proto" {
		proto.Unmarshal(ser, &b)
	}

	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)
	cpu := CPUTime() - cpuStart
	var data []byte
	if payload == "json" {
		res := &Res{
			Cpu:     cpu.Microseconds(),
			Mem:     memAfter.TotalAlloc - memBefore.TotalAlloc,
			Arrived: arrive.UnixMicro(),
		}
		data, _ = json.Marshal(res)
	} else if payload == "proto" {
		res := &Result{
			Cpu:     cpu.Microseconds(),
			Mem:     memAfter.TotalAlloc - memBefore.TotalAlloc,
			Arrived: arrive.UnixMicro(),
		}
		data, _ = proto.Marshal(res)
	}

	fmt.Fprint(w, string(data))
}

func CPUTime() time.Duration {
	var ru syscall.Rusage
	syscall.Getrusage(syscall.RUSAGE_SELF, &ru)
	return time.Duration(ru.Utime.Sec)*time.Second +
		time.Duration(ru.Utime.Usec)*time.Microsecond
}
