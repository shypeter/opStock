package opStock

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

type APIResp struct {
	MsgArray []StockInfo `json:"msgArray"`
}

type StockInfo struct {
	Name        string `json:"n"`
	Type        string `json:"ex"`
	ID          string `json:"ch"`
	Open        string `json:"o"`
	Ytdy        string `json:"y"`
	High        string `json:"h"`
	Low         string `json:"l"`
	Max         string `json:"u"`
	Min         string `json:"w"`
	TradeMoney  string `json:"z"`
	TradeVolume string `json:"tv"`
	TradeTime   string `json:"t"`
	SellVolume  string `json:"a"`
	BuyVolume   string `json:"b"`
}

func CallAPI(ctx context.Context, stocksStr string, rdb *redis.Client) error {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://mis.twse.com.tw/stock/api/getStockInfo.jsp?ex_ch="+stocksStr+"&json=1&delay=0", nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	response, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("API call failed: %w", err)
	}
	defer response.Body.Close()

	var apiResp APIResp
	err = json.NewDecoder(response.Body).Decode(&apiResp)
	if err != nil {
		return fmt.Errorf("json decode failed: %w", err)
	}

	return updateRedis(ctx, rdb, apiResp.MsgArray)
}

func updateRedis(ctx context.Context, rdb *redis.Client, stockInfos []StockInfo) error {
	pipe := rdb.Pipeline()
	for _, stockInfo := range stockInfos {
		if stockInfo.TradeMoney != "-" {
			stockData := map[string]interface{}{
				"ID":          stockInfo.ID,
				"TradeMoney":  stockInfo.TradeMoney,
				"TradeVolume": stockInfo.TradeVolume,
				"TradeTime":   stockInfo.TradeTime,
				"BuyVolume":   stockInfo.BuyVolume,
				"SellVolume":  stockInfo.SellVolume,
			}
			key := stockInfo.Type + "_" + stockInfo.ID
			pipe.HSet(ctx, key, stockData)
		}
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("Redis pipeline execution failed: %w", err)
	}

	return nil
}
