package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Gamma API 基础 URL
const gammaAPI = "https://gamma-api.polymarket.com"

// Market 结构体：解析 API 返回的市场数据
type Market struct {
	ID            string `json:"id"`
	Question      string `json:"question"`
	Type          string `json:"type"`
	Volume        string `json:"volume"`        // 字符串形式
	Outcomes      string `json:"outcomes"`      // JSON 字符串，如 "[\"Yes\", \"No\"]"
	OutcomePrices string `json:"outcomePrices"` // JSON 字符串，如 "[\"0.0345\", \"0.9655\"]"
}

// ArbitrageOpportunity 结构体：套利机会
type ArbitrageOpportunity struct {
	MarketID            string
	Question            string
	TotalCost           float64
	ArbitragePercent    float64
	SuggestedInvestment float64
	SharesToBuy         map[string]float64
}

func proxyClient() (*http.Client, error) {
	// 配置 ClashX HTTP 代理
	proxyURL, err := url.Parse("http://127.0.0.1:7890")
	if err != nil {
		return nil, fmt.Errorf("解析代理 URL 失败: %v", err)
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 10 * time.Second,
	}
	return client, nil
}

func httpProxy(client *http.Client, fetchUrl string) (*http.Response, error) {
	req, err := http.NewRequest("GET", fetchUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API 请求失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 返回错误: %d", resp.StatusCode)
	}
	return resp, nil
}

// fetchMarkets 获取所有活跃市场数据
func fetchMarkets() ([]Market, error) {
	var allMarkets []Market
	offset := 0
	limit := 100

	// 配置 ClashX HTTP 代理
	client, err := proxyClient()
	if err != nil {
		return nil, fmt.Errorf("创建代理失败: %v", err)
	}

	for {
		endDateMin := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
		endDateMax := time.Now().AddDate(0, 0, 49).Format("2006-01-02")

		fetchUrl := fmt.Sprintf("%s/markets?active=true&order=volume&ascending=false&end_date_min=%s&end_date_max=%s&limit=%d&offset=%d",
			gammaAPI, endDateMin, endDateMax, limit, offset)

		resp, err := httpProxy(client, fetchUrl)
		if err != nil {
			return nil, fmt.Errorf("API 请求失败 (offset=%d): %v", offset, err)
		}
		defer resp.Body.Close()

		var markets []Market
		if err := json.NewDecoder(resp.Body).Decode(&markets); err != nil {
			return nil, fmt.Errorf("JSON 解析失败 (offset=%d): %v", offset, err)
		}

		// 累积市场数据
		allMarkets = append(allMarkets, markets...)

		fmt.Printf("发现 %d 个活跃市场\n", len(markets))
		var opportunities []ArbitrageOpportunity
		for _, market := range markets {
			if opp := calculateArbitrage(market); opp != nil {
				opportunities = append(opportunities, *opp)
			}
		}

		if len(opportunities) == 0 {
			fmt.Println("暂无套利机会（总概率成本 >= 98%）。")
		} else {
			fmt.Printf("发现 %d 个套利机会！\n\n", len(opportunities))
			for i, opp := range opportunities {
				fmt.Printf("机会 %d:\n", i+1)
				fmt.Printf("  市场: %s\n", opp.Question)
				fmt.Printf("  市场 ID: %s\n", opp.MarketID)
				fmt.Printf("  总成本: $%.4f USDC (预期利润: %.2f%%)\n", opp.TotalCost, opp.ArbitragePercent)
				fmt.Printf("  建议投资: $%.0f USDC\n", opp.SuggestedInvestment)
				fmt.Println("  买入建议:")
				for outcome, shares := range opp.SharesToBuy {
					fmt.Printf("    - %s: %.2f 股 (价格: $%.4f)\n", outcome, shares, opp.SuggestedInvestment/float64(len(opp.SharesToBuy))/shares)
				}
				fmt.Print("\n" + "==================================================" + "\n")
			}
		}

		// 递增 offset，获取下一页
		offset += limit
		fmt.Printf("已分析 %d 个市场，继续获取 offset=%d...\n\n", len(allMarkets), offset)

		// 如果返回的市场数量 < limit，说明已到最后一页
		if len(markets) < limit {
			break
		}
	}

	return allMarkets, nil
}

// calculateArbitrage 计算单个市场的套利机会
func calculateArbitrage(market Market) *ArbitrageOpportunity {
	// 解析 outcomes JSON 字符串
	var outcomes []string
	if err := json.Unmarshal([]byte(market.Outcomes), &outcomes); err != nil {
		fmt.Printf("解析 outcomes 失败（市场: %s）: %v\n", market.Question, err)
		return nil
	}

	// 解析 outcomePrices JSON 字符串
	var priceStrs []string
	if err := json.Unmarshal([]byte(market.OutcomePrices), &priceStrs); err != nil {
		fmt.Printf("解析 outcomePrices 失败（市场: %s）: %v\n", market.Question, err)
		return nil
	}

	// 转换价格为 float64
	prices := make([]float64, len(priceStrs))
	for i, priceStr := range priceStrs {
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			fmt.Printf("转换价格失败（市场: %s, 结果: %s）: %v\n", market.Question, outcomes[i], err)
			return nil
		}
		prices[i] = price
	}

	// 过滤无效市场
	volume, err := strconv.ParseFloat(market.Volume, 64)
	if err != nil || len(outcomes) < 2 || market.Type == "BINARY" || volume < 10000 {
		return nil // 过滤二元市场、低量市场或解析失败
	}

	// 计算总成本
	totalCost := 0.0
	for _, price := range prices {
		totalCost += price
	}

	if totalCost >= 1.0 {
		return nil // 无套利机会
	}

	// 计算套利百分比
	arbitragePct := (1.0 - totalCost) * 100
	if arbitragePct < 2.0 { // 至少 2% 利润
		return nil
	}

	// 计算建议买入股份（投资 1000 USDC）
	investment := 1000.0
	sharesToBuy := make(map[string]float64)
	for i, outcome := range outcomes {
		if prices[i] > 0 {
			sharesToBuy[outcome] = investment / float64(len(outcomes)) / prices[i]
		}
	}

	return &ArbitrageOpportunity{
		MarketID:            market.ID,
		Question:            market.Question,
		TotalCost:           totalCost,
		ArbitragePercent:    arbitragePct,
		SuggestedInvestment: investment,
		SharesToBuy:         sharesToBuy,
	}
}

func main() {
	fmt.Println("=== Polymarket 套利监控程序 ===")
	fmt.Printf("当前时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Print("监控活跃市场（结束日期: 7-49 天内，volume 高优先）..." + "\n")
	fmt.Println()

	markets, err := fetchMarkets()
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		fmt.Println("检查网络或 API 状态。")
		return
	}
	fmt.Printf("共分析 %d 个活跃市场\n", len(markets))
}
