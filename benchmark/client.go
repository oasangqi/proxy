package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"

	log "github.com/golang/glog"
	"golang.org/x/net/websocket"
)

var (
	origin     = "http://localhost/"
	aliveCount int64
)

type Proto struct {
	Cmd  int64 `json:"cmd"`
	Room int64 `json:"room"`
	Seq  int64 `json:"seq"`
	Uid  int64 `json:"uid"`
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	begin, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}
	num, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}
	go result()
	for i := begin; i < begin+num; i++ {
		go client(int64(i))
	}
	// signal
	var exit chan bool
	<-exit
}

func result() {
	var (
		interval = int64(5)
	)
	for {
		nowAlive := atomic.LoadInt64(&aliveCount)
		fmt.Println(fmt.Sprintf("%s total:%s alive:%d", time.Now().Format("2006-01-02 15:04:05"), os.Args[2], nowAlive))
		time.Sleep(time.Second * time.Duration(interval))
		log.Flush()
	}
}

func client(mid int64) {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
	startClient(mid)
}

func startClient(uid int64) {
	atomic.AddInt64(&aliveCount, 1)
	quit := make(chan bool, 1)
	defer func() {
		close(quit)
		atomic.AddInt64(&aliveCount, -1)
	}()

	// 连接
	ws, err := websocket.Dial(os.Args[3], "", origin)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		hbProto := &Proto{
			Cmd:  1,
			Room: 38023802, //3286833,
			Seq:  0,
			Uid:  uid,
		}
		for {
			hbProto.Seq++
			msg, _ := json.Marshal(hbProto)
			if _, err := ws.Write(msg); err != nil {
				log.Errorf("uid:%d tcpWriteProto() error(%v)", uid, err)
				return
			}
			time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
			select {
			case <-quit:
				return
			default:
			}
		}
	}()
	var (
		lastSeq   int64 = 0
		recv      []byte
		recvproto Proto
		n         int
	)
	for {
		recv = make([]byte, 256)
		if n, err = ws.Read(recv); err != nil {
			log.Errorf("uid:%d ws.Read() error(%v)", uid, err)
			quit <- true
			return
		}
		if err := json.Unmarshal(recv[:n], &recvproto); err != nil {
			log.Errorf("uid:%d read invalid msg(%s) error(%v)", uid, recv[:n], err)
			quit <- true
			return
		}
		lastSeq++
		if recvproto.Seq != lastSeq {
			log.Errorf("uid:%d lastReq(%d) req(%d) proto(%v) n(%d) recv(%s)", uid, lastSeq, recvproto.Seq, recvproto, n, recv[:n])
			quit <- true
			return
		}
	}
}
