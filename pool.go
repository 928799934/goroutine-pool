package pool

import (
	"log"
	"sync"
)

type channelData struct {
	fn   func(chan bool, ...interface{})
	data []interface{}
}

type Runtime struct {
	channel chan *channelData
	closing chan bool
}

func newRuntime() *Runtime {
	objRuntime := &Runtime{
		make(chan *channelData, 1),
		make(chan bool, 1),
	}
	wg.Add(1)
	go objRuntime.Run()
	return objRuntime
}

func (this *Runtime) Run() {
	defer wg.Done()
	for {
		data, ok := <-this.channel
		if !ok {
			break
		}
		data.fn(this.closing, data.data...)
		mutex.Lock()
		if !readyClosed {
			mutex.Unlock()
			clean(this)
			pool <- this
			log.Println("<<<")
			continue
		}
		mutex.Unlock()
	}
}

func (this *Runtime) Close() {
	close(this.channel)
	close(this.closing)
}

var (
	mutex       sync.Mutex
	wg          sync.WaitGroup
	readyClosed bool
	poolSize    int
	pool        chan *Runtime
	running     []*Runtime
)

func Init(size int) {
	poolSize = size
	pool = make(chan *Runtime, poolSize)
	for i := 0; i < poolSize; i++ {
		pool <- newRuntime()
	}
}

func Close() {
	log.Println(len(pool))
	log.Println(len(running))
	mutex.Lock()
	if readyClosed {
		mutex.Unlock()
		return
	}
	readyClosed = true
	mutex.Unlock()
	close(pool)
	clean()
	for {
		runtime, ok := <-pool
		if !ok {
			break
		}
		runtime.Close()
	}
	wg.Wait()
}

func Run(fn func(chan bool, ...interface{}), data ...interface{}) bool {
	mutex.Lock()
	if readyClosed {
		mutex.Unlock()
		return false
	}
	mutex.Unlock()
	runtime, ok := <-pool
	if !ok {
		return false
	}
	runtime.channel <- &channelData{fn, data}
	mutex.Lock()
	running = append(running, runtime)
	mutex.Unlock()
	log.Println(">>>")
	return true
}

func clean(args ...interface{}) {
	mutex.Lock()
	var this *Runtime
	if len(args) > 0 {
		this = args[0].(*Runtime)
	}
	tmp_running := []*Runtime{}
	for _, runtime := range running {
		if this == nil {
			runtime.Close()
			continue
		}
		if runtime == this {
			continue
		}
		tmp_running = append(tmp_running, runtime)
	}
	running = tmp_running
	mutex.Unlock()
}
