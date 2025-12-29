package calculator

import (
	"fmt"
	"math"

	"github.com/sshetty/finEdSkywalker/internal/finance"
)

// DCFInput represents user-provided or default inputs for DCF
type DCFInput struct {
	RevenueGrowthRate  *float64 // Optional: if nil, use defaults/consensus
	ProfitMargin       *float64 // Optional
	FCFMargin          *float64 // Optional
	DiscountRate       *float64 // Optional (WACC or required return)
	TerminalGrowthRate *float64 // Optional
}

// CalculateDCF performs a Discounted Cash Flow valuation
func CalculateDCF(companyData *finance.CompanyData, input *DCFInput) (*finance.ValuationResult, error) {
	if companyData.LatestFinancials == nil {
		return nil, fmt.Errorf("no financial data available for DCF calculation")
	}
	
	if companyData.SharesOutstanding <= 0 {
		return nil, fmt.Errorf("shares outstanding not available")
	}
	
	// Build assumptions with defaults or user inputs
	assumptions := buildAssumptions(companyData, input)
	
	// Project cash flows for 5 years
	projections := projectCashFlows(companyData.LatestFinancials, assumptions)
	
	// Calculate terminal value
	terminalValue := calculateTerminalValue(projections, assumptions)
	
	// Calculate present value of all cash flows
	pvOfCashFlows := 0.0
	for i := range projections {
		pvOfCashFlows += projections[i].PresentValue
	}
	
	// Discount terminal value to present
	terminalDiscountFactor := math.Pow(1+assumptions.DiscountRate, float64(assumptions.ProjectionYears))
	pvOfTerminalValue := terminalValue / terminalDiscountFactor
	
	// Enterprise Value = PV of cash flows + PV of terminal value
	enterpriseValue := pvOfCashFlows + pvOfTerminalValue
	
	// For simplicity, assume Enterprise Value ≈ Equity Value (ignoring debt/cash adjustments)
	// In production, you'd subtract net debt and add cash
	equityValue := enterpriseValue
	
	// Calculate fair value per share
	fairValuePerShare := equityValue / (companyData.SharesOutstanding * 1_000_000) // Shares in millions
	
	// Calculate upside/downside
	currentPrice := 0.0
	if companyData.Quote != nil {
		currentPrice = companyData.Quote.CurrentPrice
	}
	
	upsidePercent := 0.0
	if currentPrice > 0 {
		upsidePercent = ((fairValuePerShare - currentPrice) / currentPrice) * 100
	}
	
	result := &finance.ValuationResult{
		FairValuePerShare: fairValuePerShare,
		CurrentPrice:      currentPrice,
		UpsidePercent:     upsidePercent,
		Model:             "DCF",
		Assumptions:       assumptions,
		Projections:       projections,
		TerminalValue:     terminalValue,
		EnterpriseValue:   enterpriseValue,
		SharesOutstanding: companyData.SharesOutstanding,
	}
	
	return result, nil
}

// buildAssumptions creates DCF assumptions with fallback logic
func buildAssumptions(data *finance.CompanyData, input *DCFInput) finance.DCFAssumptions {
	assumptions := finance.DCFAssumptions{
		ProjectionYears: 5,
		Source:          "defaults",
	}
	
	// Try to get from user input first
	if input != nil {
		if input.RevenueGrowthRate != nil {
			assumptions.RevenueGrowthRate = *input.RevenueGrowthRate
			assumptions.Source = "user_input"
		}
		if input.ProfitMargin != nil {
			assumptions.ProfitMargin = *input.ProfitMargin
		}
		if input.FCFMargin != nil {
			assumptions.FCFMargin = *input.FCFMargin
		}
		if input.DiscountRate != nil {
			assumptions.DiscountRate = *input.DiscountRate
		}
		if input.TerminalGrowthRate != nil {
			assumptions.TerminalGrowthRate = *input.TerminalGrowthRate
		}
	}
	
	// Apply defaults for missing values
	if assumptions.RevenueGrowthRate == 0 {
		// TODO: In future, fetch analyst consensus from Finnhub
		// For now, use reasonable default based on company size
		assumptions.RevenueGrowthRate = 0.08 // 8% default growth
		if assumptions.Source != "user_input" {
			assumptions.Source = "defaults"
		}
	}
	
	if assumptions.ProfitMargin == 0 {
		// Calculate historical profit margin if available
		if data.LatestFinancials != nil && data.LatestFinancials.Revenue > 0 {
			historicalMargin := data.LatestFinancials.NetIncome / data.LatestFinancials.Revenue
			if historicalMargin > 0 && historicalMargin < 1 {
				assumptions.ProfitMargin = historicalMargin
			} else {
				assumptions.ProfitMargin = 0.15 // 15% default
			}
		} else {
			assumptions.ProfitMargin = 0.15 // 15% default
		}
	}
	
	if assumptions.FCFMargin == 0 {
		// Calculate historical FCF margin if available
		if data.LatestFinancials != nil && data.LatestFinancials.Revenue > 0 && data.LatestFinancials.FreeCashFlow > 0 {
			historicalFCFMargin := data.LatestFinancials.FreeCashFlow / data.LatestFinancials.Revenue
			if historicalFCFMargin > 0 && historicalFCFMargin < 1 {
				assumptions.FCFMargin = historicalFCFMargin
			} else {
				assumptions.FCFMargin = 0.12 // 12% default
			}
		} else {
			assumptions.FCFMargin = 0.12 // 12% default
		}
	}
	
	if assumptions.DiscountRate == 0 {
		assumptions.DiscountRate = 0.10 // 10% default required return
	}
	
	if assumptions.TerminalGrowthRate == 0 {
		assumptions.TerminalGrowthRate = 0.025 // 2.5% perpetual growth
	}
	
	return assumptions
}

// projectCashFlows generates 5-year cash flow projections
func projectCashFlows(financials *finance.FinancialStatement, assumptions finance.DCFAssumptions) []finance.DCFProjection {
	projections := make([]finance.DCFProjection, assumptions.ProjectionYears)
	
	currentRevenue := financials.Revenue
	
	for i := 0; i < assumptions.ProjectionYears; i++ {
		year := i + 1
		
		// Project revenue with growth rate
		projectedRevenue := currentRevenue * math.Pow(1+assumptions.RevenueGrowthRate, float64(year))
		
		// Project net income
		projectedNetIncome := projectedRevenue * assumptions.ProfitMargin
		
		// Project free cash flow
		projectedFCF := projectedRevenue * assumptions.FCFMargin
		
		// Calculate discount factor
		discountFactor := math.Pow(1+assumptions.DiscountRate, float64(year))
		
		// Calculate present value
		presentValue := projectedFCF / discountFactor
		
		projections[i] = finance.DCFProjection{
			Year:           year,
			Revenue:        projectedRevenue,
			NetIncome:      projectedNetIncome,
			FreeCashFlow:   projectedFCF,
			DiscountFactor: discountFactor,
			PresentValue:   presentValue,
		}
	}
	
	return projections
}

// calculateTerminalValue calculates the terminal value using perpetuity growth model
func calculateTerminalValue(projections []finance.DCFProjection, assumptions finance.DCFAssumptions) float64 {
	if len(projections) == 0 {
		return 0
	}
	
	// Get the last year's FCF
	lastYearFCF := projections[len(projections)-1].FreeCashFlow
	
	// Terminal FCF (year after projections, with terminal growth)
	terminalFCF := lastYearFCF * (1 + assumptions.TerminalGrowthRate)
	
	// Terminal Value = Terminal FCF / (Discount Rate - Terminal Growth Rate)
	terminalValue := terminalFCF / (assumptions.DiscountRate - assumptions.TerminalGrowthRate)
	
	return terminalValue
}

// CalculateFairValueSimple is a simplified valuation method based on P/E multiples
func CalculateFairValueSimple(companyData *finance.CompanyData, targetPE float64) (float64, error) {
	if companyData.LatestFinancials == nil || companyData.SharesOutstanding <= 0 {
		return 0, fmt.Errorf("insufficient data for simple valuation")
	}
	
	if targetPE == 0 {
		targetPE = 15.0 // Default target P/E
	}
	
	eps := companyData.LatestFinancials.NetIncome / (companyData.SharesOutstanding * 1_000_000)
	fairValue := eps * targetPE
	
	return fairValue, nil
}

// GetAnalystConsensus would fetch analyst estimates from Finnhub
// For MVP, this returns nil to use defaults
func GetAnalystConsensus(ticker string) *DCFInput {
	// TODO: Implement Finnhub analyst estimates API call
	// Endpoint: /stock/price-target and /stock/recommendation
	return nil
}

