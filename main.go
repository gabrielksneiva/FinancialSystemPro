// @title FinancialSystemPro API
// @version 1.0
// @description API do FinancialSystemPro
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Digite "Bearer {seu_token}"
package main

import "financial-system-pro/api"

func main() {
	api.Start()
}
