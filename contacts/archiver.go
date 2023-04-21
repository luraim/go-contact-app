package contacts

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Archiver struct {
	archiveStatus   string
	archiveProgress float64
	thread          *sync.WaitGroup
	mutex           sync.Mutex
}

func NewArchiver() *Archiver {
	return &Archiver{
		archiveStatus: "Waiting",
	}
}

func (a *Archiver) Status() string {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	return a.archiveStatus
}

func (a *Archiver) Progress() float64 {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	return a.archiveProgress
}

func (a *Archiver) ProgressPercentage() float64 {
	return a.Progress() * 100
}

func (a *Archiver) Run() {
	a.mutex.Lock()
	if a.archiveStatus == "Waiting" {
		a.archiveStatus = "Running"
		a.archiveProgress = 0
		a.thread = &sync.WaitGroup{}
		a.thread.Add(1)
		go a.runImpl()
	}
	a.mutex.Unlock()
}

func (a *Archiver) runImpl() {
	defer a.thread.Done()
	for i := 0; i < 10; i++ {
		time.Sleep(time.Duration(rand.Float64()) * time.Second)
		a.mutex.Lock()
		if a.archiveStatus != "Running" {
			a.mutex.Unlock()
			return
		}
		a.archiveProgress = float64(i+1) / 10.0
		fmt.Printf("Here... %.2f\n", a.archiveProgress)
		a.mutex.Unlock()
	}
	time.Sleep(1 * time.Second)
	a.mutex.Lock()
	if a.archiveStatus != "Running" {
		a.mutex.Unlock()
		return
	}
	a.archiveStatus = "Complete"
	a.mutex.Unlock()
}

func (a *Archiver) ArchiveFile() string {
	return dbJsonFileName
}

func (a *Archiver) Reset() {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.archiveStatus = "Waiting"
}
