package main

import "sync"

type PlayerState struct {
	mu    sync.Mutex
	Coins int
	Stars int
}

var playerStore = &PlayerState{
	Coins: 100, // starter coins
	Stars: 0,
}

func (p *PlayerState) Get() (coins int, stars int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.Coins, p.Stars
}

func (p *PlayerState) CanAfford(cost int) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.Coins >= cost
}

func (p *PlayerState) ApplyPurchase(cost int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Coins -= cost
	p.Stars += 1
}
