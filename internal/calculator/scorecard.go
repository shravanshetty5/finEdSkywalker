package calculator

import (
	"fmt"
	"math"

	"github.com/sshetty/finEdSkywalker/internal/finance"
)

// CalculateScorecard generates the Big 5 fundamental metrics scorecard
func CalculateScorecard(companyData *finance.CompanyData) *finance.FundamentalScorecard {
	scorecard := &finance.FundamentalScorecard{}
	
	// 1. P/E Ratio
	scorecard.PERatio = calculatePERatio(companyData)
	
	// 2. Debt-to-Equity Ratio
	scorecard.DebtToEquity = calculateDebtToEquity(companyData)
	
	// 3. FCF Yield
	scorecard.FCFYield = calculateFCFYield(companyData)
	
	// 4. PEG Ratio
	scorecard.PEGRatio = calculatePEGRatio(companyData)
	
	// 5. ROE (Return on Equity)
	scorecard.ROE = calculateROE(companyData)
	
	// Calculate overall score
	scorecard.OverallScore, scorecard.Summary = calculateOverallScore(scorecard)
	
	return scorecard
}

// calculatePERatio calculates and rates the P/E ratio
func calculatePERatio(data *finance.CompanyData) finance.FundamentalMetric {
	metric := finance.FundamentalMetric{
		Available: false,
		Rating:    finance.RatingNA,
	}
	
	if data.Quote == nil || data.LatestFinancials == nil {
		metric.Message = "Insufficient data to calculate P/E ratio"
		return metric
	}
	
	if data.LatestFinancials.NetIncome <= 0 || data.SharesOutstanding <= 0 {
		metric.Message = "Company has negative or zero earnings"
		metric.Rating = finance.RatingRed
		return metric
	}
	
	// Calculate P/E = Price / EPS
	// EPS = Net Income / Shares Outstanding
	eps := data.LatestFinancials.NetIncome / (data.SharesOutstanding * 1_000_000) // Shares in millions
	peRatio := data.Quote.CurrentPrice / eps
	
	metric.Current = peRatio
	metric.Available = true
	
	// Compare against 5-year average if available
	if data.HistoricalData != nil && data.HistoricalData.PERatioAvg5Year > 0 {
		fiveYearAvg := data.HistoricalData.PERatioAvg5Year
		metric.FiveYearAvg = &fiveYearAvg
		
		// Rating logic
		if peRatio < fiveYearAvg*0.9 {
			metric.Rating = finance.RatingGreen
			metric.Message = fmt.Sprintf("P/E (%.2f) is below 5-year average (%.2f) - potentially undervalued", peRatio, fiveYearAvg)
		} else if peRatio > fiveYearAvg*1.2 {
			metric.Rating = finance.RatingRed
			metric.Message = fmt.Sprintf("P/E (%.2f) is above 5-year average (%.2f) - potentially overvalued", peRatio, fiveYearAvg)
		} else {
			metric.Rating = finance.RatingYellow
			metric.Message = fmt.Sprintf("P/E (%.2f) is near 5-year average (%.2f)", peRatio, fiveYearAvg)
		}
	} else {
		// No historical data, use general benchmarks
		if peRatio < 15 {
			metric.Rating = finance.RatingGreen
			metric.Message = fmt.Sprintf("P/E of %.2f suggests good value", peRatio)
		} else if peRatio > 30 {
			metric.Rating = finance.RatingRed
			metric.Message = fmt.Sprintf("P/E of %.2f is relatively high", peRatio)
		} else {
			metric.Rating = finance.RatingYellow
			metric.Message = fmt.Sprintf("P/E of %.2f is moderate", peRatio)
		}
	}
	
	return metric
}

// calculateDebtToEquity calculates and rates the Debt-to-Equity ratio
func calculateDebtToEquity(data *finance.CompanyData) finance.FundamentalMetric {
	metric := finance.FundamentalMetric{
		Available: false,
		Rating:    finance.RatingNA,
	}
	
	if data.LatestFinancials == nil {
		metric.Message = "No financial data available"
		return metric
	}
	
	if data.LatestFinancials.ShareholdersEquity <= 0 {
		metric.Message = "Company has negative or zero equity"
		metric.Rating = finance.RatingRed
		return metric
	}
	
	debtToEquity := data.LatestFinancials.TotalDebt / data.LatestFinancials.ShareholdersEquity
	metric.Current = debtToEquity
	metric.Available = true
	
	// Rating logic (lower is better for long-term safety)
	if debtToEquity < 0.5 {
		metric.Rating = finance.RatingGreen
		metric.Message = fmt.Sprintf("Excellent debt levels (%.2f) - very safe", debtToEquity)
	} else if debtToEquity < 1.0 {
		metric.Rating = finance.RatingYellow
		metric.Message = fmt.Sprintf("Moderate debt levels (%.2f) - acceptable", debtToEquity)
	} else if debtToEquity < 2.0 {
		metric.Rating = finance.RatingRed
		metric.Message = fmt.Sprintf("High debt levels (%.2f) - risky", debtToEquity)
	} else {
		metric.Rating = finance.RatingRed
		metric.Message = fmt.Sprintf("Very high debt levels (%.2f) - concerning", debtToEquity)
	}
	
	return metric
}

// calculateFCFYield calculates and rates the Free Cash Flow Yield
func calculateFCFYield(data *finance.CompanyData) finance.FundamentalMetric {
	metric := finance.FundamentalMetric{
		Available: false,
		Rating:    finance.RatingNA,
	}
	
	if data.Quote == nil || data.LatestFinancials == nil {
		metric.Message = "Insufficient data to calculate FCF Yield"
		return metric
	}
	
	if data.Quote.MarketCap <= 0 {
		metric.Message = "Market cap not available"
		return metric
	}
	
	// Calculate FCF Yield = Free Cash Flow / Market Cap
	fcfYield := (data.LatestFinancials.FreeCashFlow / data.Quote.MarketCap) * 100 // As percentage
	
	metric.Current = fcfYield
	metric.Available = true
	
	// Rating logic (higher is better)
	if fcfYield > 8 {
		metric.Rating = finance.RatingGreen
		metric.Message = fmt.Sprintf("Excellent FCF yield (%.2f%%) - strong cash generation", fcfYield)
	} else if fcfYield > 4 {
		metric.Rating = finance.RatingYellow
		metric.Message = fmt.Sprintf("Good FCF yield (%.2f%%)", fcfYield)
	} else if fcfYield > 0 {
		metric.Rating = finance.RatingRed
		metric.Message = fmt.Sprintf("Low FCF yield (%.2f%%) - limited cash generation", fcfYield)
	} else {
		metric.Rating = finance.RatingRed
		metric.Message = fmt.Sprintf("Negative FCF yield (%.2f%%) - burning cash", fcfYield)
	}
	
	return metric
}

// calculatePEGRatio calculates and rates the PEG ratio
func calculatePEGRatio(data *finance.CompanyData) finance.FundamentalMetric {
	metric := finance.FundamentalMetric{
		Available: false,
		Rating:    finance.RatingNA,
	}
	
	// PEG requires P/E ratio and growth rate
	peMetric := calculatePERatio(data)
	if !peMetric.Available {
		metric.Message = "P/E ratio not available"
		return metric
	}
	
	// Use a default growth rate estimate (8%) if not provided
	// In production, this would come from analyst estimates
	growthRate := 8.0 // 8% default
	
	if growthRate <= 0 {
		metric.Message = "Growth rate not available or negative"
		return metric
	}
	
	// PEG = PE / Growth Rate
	pegRatio := peMetric.Current / growthRate
	
	metric.Current = pegRatio
	metric.Available = true
	
	// Rating logic (lower is better, < 1 is undervalued)
	if pegRatio < 1.0 {
		metric.Rating = finance.RatingGreen
		metric.Message = fmt.Sprintf("PEG of %.2f suggests undervalued relative to growth", pegRatio)
	} else if pegRatio < 1.5 {
		metric.Rating = finance.RatingYellow
		metric.Message = fmt.Sprintf("PEG of %.2f is fairly valued", pegRatio)
	} else {
		metric.Rating = finance.RatingRed
		metric.Message = fmt.Sprintf("PEG of %.2f suggests overvalued relative to growth", pegRatio)
	}
	
	metric.Message += fmt.Sprintf(" (assuming %.0f%% growth)", growthRate)
	
	return metric
}

// calculateROE calculates and rates the Return on Equity
func calculateROE(data *finance.CompanyData) finance.FundamentalMetric {
	metric := finance.FundamentalMetric{
		Available: false,
		Rating:    finance.RatingNA,
	}
	
	if data.LatestFinancials == nil {
		metric.Message = "No financial data available"
		return metric
	}
	
	if data.LatestFinancials.ShareholdersEquity <= 0 {
		metric.Message = "Company has negative or zero equity"
		metric.Rating = finance.RatingRed
		return metric
	}
	
	// ROE = Net Income / Shareholders' Equity (as percentage)
	roe := (data.LatestFinancials.NetIncome / data.LatestFinancials.ShareholdersEquity) * 100
	
	metric.Current = roe
	metric.Available = true
	
	// Rating logic (higher is better - indicates management efficiency)
	if roe > 20 {
		metric.Rating = finance.RatingGreen
		metric.Message = fmt.Sprintf("Excellent ROE (%.2f%%) - highly efficient management", roe)
	} else if roe > 15 {
		metric.Rating = finance.RatingYellow
		metric.Message = fmt.Sprintf("Good ROE (%.2f%%) - solid management", roe)
	} else if roe > 0 {
		metric.Rating = finance.RatingRed
		metric.Message = fmt.Sprintf("Low ROE (%.2f%%) - poor capital efficiency", roe)
	} else {
		metric.Rating = finance.RatingRed
		metric.Message = fmt.Sprintf("Negative ROE (%.2f%%) - losing money", roe)
	}
	
	return metric
}

// calculateOverallScore summarizes the scorecard
func calculateOverallScore(scorecard *finance.FundamentalScorecard) (string, string) {
	greenCount := 0
	yellowCount := 0
	redCount := 0
	totalMetrics := 0
	
	metrics := []finance.FundamentalMetric{
		scorecard.PERatio,
		scorecard.DebtToEquity,
		scorecard.FCFYield,
		scorecard.PEGRatio,
		scorecard.ROE,
	}
	
	for _, metric := range metrics {
		if metric.Available {
			totalMetrics++
			switch metric.Rating {
			case finance.RatingGreen:
				greenCount++
			case finance.RatingYellow:
				yellowCount++
			case finance.RatingRed:
				redCount++
			}
		}
	}
	
	if totalMetrics == 0 {
		return "0/5 metrics available", "Insufficient data for analysis"
	}
	
	healthyCount := greenCount + yellowCount
	score := fmt.Sprintf("%d/%d metrics healthy", healthyCount, totalMetrics)
	
	var summary string
	percentage := float64(greenCount) / float64(totalMetrics) * 100
	
	if percentage >= 60 {
		summary = "Strong fundamentals - Good investment candidate"
	} else if percentage >= 40 {
		summary = "Mixed fundamentals - Proceed with caution"
	} else {
		summary = "Weak fundamentals - High risk"
	}
	
	if redCount > totalMetrics/2 {
		summary = "Concerning fundamentals - Avoid or investigate further"
	}
	
	return score, summary
}

// Helper function to safely get float value
func safeFloat(val float64, defaultVal float64) float64 {
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return defaultVal
	}
	return val
}

