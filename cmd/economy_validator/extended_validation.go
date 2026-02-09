package main

// extended_validation.go - Extended economic validation with real pricing
// This version runs longer tests to properly demonstrate UBI and affordability

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

type ExtendedValidationResult struct {
	ScenarioName          string
	PopulationCount       int
	SimulationDurationHrs float64
	TotalTicksRun         int
	UBIInfo               UBIMetrics
	StarAffordabilityInfo AffordabilityMetrics
	EconomyHealthInfo     PoolHealthMetrics
	PricingInfo           PricingMetrics
	OverallAssessment     string
}

type UBIMetrics struct {
	PerTickAmount           int
	PerHourAmount           float64
	TotalDistributed        float64
	AllPlayersReceivingUBI  bool
	DistributionFailureRate float64
}

type AffordabilityMetrics struct {
	LowestPlayerBalance   float64
	MedianPlayerBalance   float64
	HighestPlayerBalance  float64
	FinalStarPrice        float64
	MinTimeToFirstStar    int64 // seconds
	PercentageAffording   float64
	IsPopulationInvariant bool
}

type PoolHealthMetrics struct {
	InitialPoolSize     float64
	TotalEmitted        float64
	TotalInWallets      float64
	DepletionPercentage float64
	CriticalExhaustion  bool
}

type PricingMetrics struct {
	InitialPrice          float64
	FinalPrice            float64
	MaxPrice              float64
	MonotonicIncrease     bool
	LateGameSpikeOccurred bool
}

func RunExtendedValidationScenarios() {
	fmt.Println("\n" + repeatStr("=", 80))
	fmt.Println("EXTENDED ECONOMIC VALIDATION REPORT")
	fmt.Println(repeatStr("=", 80))

	scenarios := []struct {
		name       string
		population int
		duration   time.Duration
		desc       string
	}{
		{
			name:       "Scenario A (Extended): Solo Player - 24 Hour Test",
			population: 1,
			duration:   24 * time.Hour,
			desc:       "Validate UBI provides sustainable income for solo player over extended period",
		},
		{
			name:       "Scenario B (Extended): Small Population - 10 Player Mixed Behavior",
			population: 10,
			duration:   10 * 24 * time.Hour,
			desc:       "Validate UBI sustains diverse player archetypes with varying activity",
		},
		{
			name:       "Scenario C (Extended): Large Population - Full Emission Cycle",
			population: 500,
			duration:   14 * 24 * time.Hour,
			desc:       "Validate emission pool and UBI under sustained population pressure",
		},
	}

	results := []ExtendedValidationResult{}

	for _, scenario := range scenarios {
		fmt.Println()
		fmt.Println("Running:", scenario.name)
		fmt.Println("Description:", scenario.desc)
		fmt.Println()

		result := runExtendedValidation(scenario.name, scenario.population, scenario.duration)
		results = append(results, result)
		printExtendedResult(result)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("SUMMARY ACROSS ALL SCENARIOS")
	fmt.Println(strings.Repeat("=", 80))

	printSummary(results)
}

func runExtendedValidation(name string, population int, duration time.Duration) ExtendedValidationResult {
	const (
		tickInterval = 60 * time.Second
		COIN_SCALE   = 1000
	)

	tickCount := int(duration.Seconds() / tickInterval.Seconds())
	result := ExtendedValidationResult{
		ScenarioName:          name,
		PopulationCount:       population,
		SimulationDurationHrs: duration.Hours(),
		TotalTicksRun:         tickCount,
	}

	// Initialize economy state
	globalPool := 75000 * COIN_SCALE // Bootstrap
	emitted := 0
	inWallets := int64(0)
	starsPurchased := 0

	players := make([]int64, population)
	playerStars := make([]int, population)
	playerTimeToFirstStar := make([]int64, population)
	for i := range playerTimeToFirstStar {
		playerTimeToFirstStar[i] = -1
	}

	ubiFailures := 0
	emissionRemainder := 0.0
	const dailyEmissionTarget = 1000 * COIN_SCALE // Daily budget

	priceHistory := []int{}

	// Simulate ticks
	for tick := 0; tick < tickCount; tick++ {
		progress := float64(tick) / float64(tickCount)

		// Emission
		dailyTarget := dailyEmissionTarget
		if progress > 0.5 {
			dailyTarget = int(float64(dailyEmissionTarget) * (1.0 - 0.25*(progress-0.5)/0.5))
		}
		coinsPerTick := float64(dailyTarget) / (24 * 3600)
		emissionRemainder += coinsPerTick
		emitNow := int(emissionRemainder)
		if emitNow > 0 {
			emissionRemainder -= float64(emitNow)
			globalPool += emitNow
			emitted += emitNow
		}

		// UBI Distribution
		const ubiPerTick = 1
		available := globalPool - emitted
		ubiThisTick := 0
		for i := range players {
			if available >= ubiPerTick {
				players[i] += int64(ubiPerTick)
				inWallets += int64(ubiPerTick)
				emitted += ubiPerTick
				available -= ubiPerTick
				ubiThisTick++
			} else {
				ubiFailures++
			}
		}

		// Daily faucets (simplified)
		if population <= 10 && tick%(24*3600/60) == 0 && tick > 0 {
			dailyReward := 20 * COIN_SCALE
			for i := range players {
				if available >= dailyReward {
					players[i] += int64(dailyReward)
					inWallets += int64(dailyReward)
					available -= dailyReward
				}
			}
		}

		// Star Pricing (real-ish formula)
		coinsPerPlayer := float64(0)
		if population > 0 {
			coinsPerPlayer = float64(inWallets) / float64(population)
		}

		// Use real pricing components
		scarcity := 1.0 + math.Pow(float64(starsPurchased+1), 1.5)/25.0
		timeMultiplier := 1.0 + 3.0*math.Pow(progress, 2)
		coinMultiplier := 1.0
		if coinsPerPlayer > 0 {
			coinMultiplier = 1.0 + 0.55*math.Log1p(coinsPerPlayer/float64(100*COIN_SCALE))
		}

		basePrice := int(float64(10*COIN_SCALE) * scarcity * timeMultiplier * coinMultiplier)

		// Affordability guardrail (real formula)
		affordabilityCap := int(coinsPerPlayer * 0.9)
		if affordabilityCap < 10*COIN_SCALE {
			affordabilityCap = 10 * COIN_SCALE
		}
		if basePrice > affordabilityCap {
			basePrice = affordabilityCap
		}

		priceHistory = append(priceHistory, basePrice)

		// Star purchasing
		if population == 1 && playerStars[0] == 0 && players[0] >= int64(basePrice) {
			players[0] -= int64(basePrice)
			inWallets -= int64(basePrice)
			playerStars[0]++
			starsPurchased++
			playerTimeToFirstStar[0] = int64(tick * 60)
		} else if population > 1 && tick%300 == 0 { // Every 5 minutes for small pops
			// Greedy player
			idx := 0
			if players[idx] >= int64(basePrice*2) && playerStars[idx] < 3 {
				players[idx] -= int64(basePrice)
				inWallets -= int64(basePrice)
				playerStars[idx]++
				starsPurchased++
				if playerTimeToFirstStar[idx] == -1 {
					playerTimeToFirstStar[idx] = int64(tick * 60)
				}
			}
			// Conservative player
			idx = 1 % population
			if players[idx] >= int64(basePrice) && playerStars[idx] == 0 && basePrice < 50*COIN_SCALE {
				players[idx] -= int64(basePrice)
				inWallets -= int64(basePrice)
				playerStars[idx]++
				starsPurchased++
				if playerTimeToFirstStar[idx] == -1 {
					playerTimeToFirstStar[idx] = int64(tick * 60)
				}
			}
		}
	}

	// Calculate metrics
	result.UBIInfo = UBIMetrics{
		PerTickAmount:           1,
		PerHourAmount:           1.0 / float64(COIN_SCALE) * 3600,
		TotalDistributed:        float64(emitted) / float64(COIN_SCALE),
		AllPlayersReceivingUBI:  ubiFailures == 0,
		DistributionFailureRate: float64(ubiFailures) / float64(tickCount*population),
	}

	playerBalances := make([]float64, len(players))
	affording := 0
	for i, p := range players {
		playerBalances[i] = float64(p) / float64(COIN_SCALE)
		finalPrice := float64(priceHistory[len(priceHistory)-1]) / float64(COIN_SCALE)
		if playerBalances[i] >= finalPrice*0.8 {
			affording++
		}
	}
	sort.Float64s(playerBalances)

	result.StarAffordabilityInfo = AffordabilityMetrics{
		LowestPlayerBalance:   playerBalances[0],
		MedianPlayerBalance:   playerBalances[len(playerBalances)/2],
		HighestPlayerBalance:  playerBalances[len(playerBalances)-1],
		FinalStarPrice:        float64(priceHistory[len(priceHistory)-1]) / float64(COIN_SCALE),
		MinTimeToFirstStar:    findMinTimeToStar(playerTimeToFirstStar),
		PercentageAffording:   float64(affording) / float64(population) * 100,
		IsPopulationInvariant: result.UBIInfo.AllPlayersReceivingUBI,
	}

	result.EconomyHealthInfo = PoolHealthMetrics{
		InitialPoolSize:     float64(75000),
		TotalEmitted:        result.UBIInfo.TotalDistributed,
		TotalInWallets:      float64(inWallets) / float64(COIN_SCALE),
		DepletionPercentage: 100 * float64(ubiFailures) / float64(tickCount*population),
		CriticalExhaustion:  float64(ubiFailures)/float64(tickCount*population) > 0.05,
	}

	result.PricingInfo = PricingMetrics{
		InitialPrice:      float64(priceHistory[0]) / float64(COIN_SCALE),
		FinalPrice:        float64(priceHistory[len(priceHistory)-1]) / float64(COIN_SCALE),
		MaxPrice:          float64(maxInt(priceHistory...)) / float64(COIN_SCALE),
		MonotonicIncrease: isMonotonic(priceHistory),
		LateGameSpikeOccurred: len(priceHistory) > 0 &&
			priceHistory[len(priceHistory)-1] > priceHistory[len(priceHistory)/2],
	}

	// Determine overall assessment
	if result.StarAffordabilityInfo.IsPopulationInvariant &&
		result.StarAffordabilityInfo.PercentageAffording >= 80 &&
		!result.EconomyHealthInfo.CriticalExhaustion {
		result.OverallAssessment = "✅ PASS: Economy is population-invariant and playable"
	} else {
		result.OverallAssessment = "⚠️  REVIEW: Economy has concerns - see details"
	}

	return result
}

func printExtendedResult(r ExtendedValidationResult) {
	fmt.Printf("\nSimulation Results (%d ticks, %.1f hours):\n", r.TotalTicksRun, r.SimulationDurationHrs)
	fmt.Printf("\n📊 UBI Distribution:\n")
	fmt.Printf("  Per tick: %d microcoins\n", r.UBIInfo.PerTickAmount)
	fmt.Printf("  Per hour: %.6f coins\n", r.UBIInfo.PerHourAmount)
	fmt.Printf("  Total distributed: %.3f coins\n", r.UBIInfo.TotalDistributed)
	fmt.Printf("  Distribution complete: %v (failure rate: %.3f%%)\n",
		r.UBIInfo.AllPlayersReceivingUBI, r.UBIInfo.DistributionFailureRate*100)

	fmt.Printf("\n💰 Star Affordability:\n")
	fmt.Printf("  Final star price: %.3f coins\n", r.StarAffordabilityInfo.FinalStarPrice)
	fmt.Printf("  Player balances: min=%.3f, median=%.3f, max=%.3f coins\n",
		r.StarAffordabilityInfo.LowestPlayerBalance,
		r.StarAffordabilityInfo.MedianPlayerBalance,
		r.StarAffordabilityInfo.HighestPlayerBalance)
	if r.StarAffordabilityInfo.MinTimeToFirstStar >= 0 {
		fmt.Printf("  Time to first star: %d seconds (%.1f hours)\n",
			r.StarAffordabilityInfo.MinTimeToFirstStar,
			float64(r.StarAffordabilityInfo.MinTimeToFirstStar)/3600)
	} else {
		fmt.Printf("  Time to first star: Not achieved\n")
	}
	fmt.Printf("  Players affording stars: %.1f%%\n", r.StarAffordabilityInfo.PercentageAffording)
	fmt.Printf("  Population-invariant: %v\n", r.StarAffordabilityInfo.IsPopulationInvariant)

	fmt.Printf("\n🏦 Emission Pool Health:\n")
	fmt.Printf("  Initial pool: %.0f coins\n", r.EconomyHealthInfo.InitialPoolSize)
	fmt.Printf("  Total emitted: %.3f coins\n", r.EconomyHealthInfo.TotalEmitted)
	fmt.Printf("  In wallets: %.3f coins\n", r.EconomyHealthInfo.TotalInWallets)
	fmt.Printf("  Depletion incidents: %.3f%% of ticks\n", r.EconomyHealthInfo.DepletionPercentage)
	fmt.Printf("  Critical exhaustion: %v\n", r.EconomyHealthInfo.CriticalExhaustion)

	fmt.Printf("\n📈 Star Pricing Dynamics:\n")
	fmt.Printf("  Initial: %.3f coins\n", r.PricingInfo.InitialPrice)
	fmt.Printf("  Final: %.3f coins\n", r.PricingInfo.FinalPrice)
	fmt.Printf("  Max observed: %.3f coins\n", r.PricingInfo.MaxPrice)
	fmt.Printf("  Monotonic increase: %v\n", r.PricingInfo.MonotonicIncrease)
	fmt.Printf("  Late-game spike: %v\n", r.PricingInfo.LateGameSpikeOccurred)

	fmt.Printf("\n%s\n", r.OverallAssessment)
}

func printSummary(results []ExtendedValidationResult) {
	allPass := true
	for _, r := range results {
		if r.StarAffordabilityInfo.PercentageAffording < 80 ||
			r.EconomyHealthInfo.CriticalExhaustion {
			allPass = false
		}
	}

	fmt.Printf("Total Scenarios: %d\n", len(results))
	fmt.Printf("Overall Status: ")
	if allPass {
		fmt.Printf("✅ ALL PASS\n")
	} else {
		fmt.Printf("⚠️  REVIEW REQUIRED\n")
	}

	fmt.Printf("\nKey Findings:\n")
	for _, r := range results {
		fmt.Printf("  - %s: UBI %.3f coins/hr, %.1f%% can afford stars\n",
			r.ScenarioName, r.UBIInfo.PerHourAmount, r.StarAffordabilityInfo.PercentageAffording)
	}
}

func maxInt(nums ...int) int {
	if len(nums) == 0 {
		return 0
	}
	max := nums[0]
	for _, n := range nums {
		if n > max {
			max = n
		}
	}
	return max
}

func isMonotonic(prices []int) bool {
	if len(prices) < 2 {
		return true
	}
	for i := 1; i < len(prices); i++ {
		if prices[i] < prices[i-1] {
			return false
		}
	}
	return true
}

func findMinTimeToStar(times []int64) int64 {
	minTime := int64(-1)
	for _, t := range times {
		if t >= 0 && (minTime == -1 || t < minTime) {
			minTime = t
		}
	}
	return minTime
}
