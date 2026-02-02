package main

import (
	"log"
	"sync"
)

type EconomyState struct {
	mu                  sync.Mutex
	globalCoinPool      int
	dailyEmissionTarget int
	emissionRemainder   float64
}

var economy = &EconomyState{
	globalCoinPool:      0,
	dailyEmissionTarget: 1000,
	emissionRemainder:   0,
}

func (e *EconomyState) emitCoins(amount int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.globalCoinPool += amount
	log.Println("Economy: emitted coins,", amount, "pool now", e.globalCoinPool)
}
