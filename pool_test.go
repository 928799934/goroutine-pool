package pool

import (
	"log"
	"testing"
	"time"
)

func callback(closed chan bool, d interface{}) {
	t := time.NewTimer(time.Second * 1)
	for {
		// 等待
		select {
		case <-closed:
			return
		case <-t.C:
		}
		// TODO 工作
		log.Println(d)
		t.Reset(time.Second * 1)
	}
}

func callback1(closed chan bool, d interface{}) {
	t := time.NewTimer(time.Second * 1)
	i := 0
	for {
		if i == 10 {
			break
		}
		// 等待
		select {
		case <-closed:
			return
		case <-t.C:
		}
		// TODO 工作
		log.Println(d)
		t.Reset(time.Second * 1)
		i++
	}
}

func TestPool(t *testing.T) {
	Init(3)
	defer func() {
		Close()
		if !Run("abc", callback) {
			log.Println("add abc fail")
		}
	}()
	go func() {
		if !Run("abc", callback) {
			log.Println("add fail")
		}
	}()
	time.Sleep(time.Millisecond * 100)
	go func() {
		if !Run("def", callback1) {
			log.Println("add def fail")
		}
	}()
	time.Sleep(time.Second * 5)
}
