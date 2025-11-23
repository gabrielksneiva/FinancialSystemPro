// @title FinancialSystemPro API
// @version 1.0
// @description API do FinancialSystemPro
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Digite "Bearer {seu_token}"
package main

import http "financial-system-pro/internal/adapters/http"

func main() {
	http.Start()
}
