package tools

import (
	"math"
)

// GenerateWRRSequence genera una secuencia WRR a partir de probabilidades
func GenerateWRRSequence(probabilities []float64) []int {
	// Paso 1: Verificar que las probabilidades sumen aproximadamente 1
	totalProb := 0.0
	for _, p := range probabilities {
		totalProb += p
	}

	if math.Abs(totalProb-1.0) > 0.001 {
		// Ajustar probabilidades para que sumen 1
		scale := 1.0 / totalProb
		for i := range probabilities {
			probabilities[i] *= scale
		}
	}

	// Paso 2: Convertir probabilidades a fracciones
	numerators := make([]int, len(probabilities))
	denominators := make([]int, len(probabilities))

	const maxDenominator = 1000
	const precision = 1e-9

	for i, prob := range probabilities {
		// Encontrar la mejor fracción
		bestDenominator := 1
		bestNumerator := int(math.Round(prob))
		bestError := math.Abs(prob - float64(bestNumerator))

		for d := 1; d <= maxDenominator; d++ {
			n := int(math.Round(prob * float64(d)))
			err := math.Abs(prob - float64(n)/float64(d))

			if err < bestError {
				bestError = err
				bestDenominator = d
				bestNumerator = n

				if err < precision {
					break
				}
			}
		}

		numerators[i] = bestNumerator
		denominators[i] = bestDenominator
	}

	// Paso 3: Encontrar denominador común (MCM)
	gcd := func(a, b int) int {
		for b != 0 {
			a, b = b, a%b
		}
		return a
	}

	lcm := func(a, b int) int {
		return a * b / gcd(a, b)
	}

	commonDenominator := 1
	for _, den := range denominators {
		commonDenominator = lcm(commonDenominator, den)
	}

	// Paso 4: Calcular pesos (numeradores escalados)
	weights := make([]int, len(probabilities))
	for i := range probabilities {
		scale := commonDenominator / denominators[i]
		weights[i] = numerators[i] * scale
	}

	// Paso 5: Simplificar pesos
	gcdAll := weights[0]
	for i := 1; i < len(weights); i++ {
		gcdAll = gcd(gcdAll, weights[i])
	}

	if gcdAll > 1 {
		for i := range weights {
			weights[i] /= gcdAll
		}
		commonDenominator /= gcdAll
	}

	// Paso 6: Generar secuencia WRR
	sequence := make([]int, 0, commonDenominator)
	currentWeights := make([]int, len(weights))

	for range commonDenominator {
		// Encontrar servidor con mayor peso actual
		maxIndex := 0
		maxWeight := currentWeights[0]

		for i := 1; i < len(weights); i++ {
			if currentWeights[i] > maxWeight {
				maxWeight = currentWeights[i]
				maxIndex = i
			}
		}

		sequence = append(sequence, maxIndex)

		// Actualizar pesos según algoritmo WRR
		currentWeights[maxIndex] -= commonDenominator
		for i := range currentWeights {
			currentWeights[i] += weights[i]
		}
	}

	return sequence
}
