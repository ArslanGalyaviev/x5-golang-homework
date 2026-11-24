package analytics

type Sale struct {
	Product          string
	Category         string
	Quantity         int
	UnitPriceKopecks int64
	Cancelled        bool
}

type Report struct {
	RevenueKopecks    int64
	UnitsSold         int
	RevenueByCategory map[string]int64
	TopProduct        string
	InvalidSales      int
}

func AnalyzeSales(sales []Sale) Report {
	report := Report{
		RevenueByCategory: make(map[string]int64),
	}
	productsUnits := make(map[string]int)
	for _, sale := range sales {
		if sale.Cancelled {
			continue
		}
		if sale.Product == "" || sale.Quantity <= 0 || sale.Category == "" || sale.UnitPriceKopecks < 0 {
			report.InvalidSales++
			continue
		}
		revenue := sale.UnitPriceKopecks * int64(sale.Quantity)

		report.RevenueKopecks += revenue
		report.UnitsSold += sale.Quantity
		report.RevenueByCategory[sale.Category] += revenue

		productsUnits[sale.Product] += sale.Quantity
	}
	topProduct := ""
	topUnits := 0
	for product, units := range productsUnits {
		if units > topUnits || (units == topUnits && product < topProduct) {
			topProduct = product
			topUnits = units
		}
	}

	report.TopProduct = topProduct
	return report
}
