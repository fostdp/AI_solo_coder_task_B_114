package handlers

import (
	"sync"
	"time"
)

var (
	analysisCounter uint64
	counterMutex    sync.Mutex
)

func generateAnalysisID() uint64 {
	counterMutex.Lock()
	defer counterMutex.Unlock()
	analysisCounter++
	return uint64(time.Now().Unix()%1000000)*10000 + analysisCounter%10000
}

func resetAnalysisCounter() {
	counterMutex.Lock()
	defer counterMutex.Unlock()
	analysisCounter = 0
}
