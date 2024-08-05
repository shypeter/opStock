package opRedis

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-redis/redis/v8"
)

func NewClient() (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("unable to connect Redis: %v", err)
	}

	return rdb, nil
}

func GetStocksInfo(ctx context.Context, rdb *redis.Client, stocks string) string {
	stocksArr := strings.Split(stocks, "|")
	res := ""
	for _, stock := range stocksArr {
		stockData, _ := rdb.HGetAll(ctx, stock).Result()
		fmt.Println(stockData)
		//"TradeMoney":
		//"TradeVolume":
		//"TradeTime":
		//"BuyVolume":
		//"SellVolume":
		res += stockData["TradeMoney"] + "_" + stockData["TradeVolume"] + "<br>"
	}
	return res
}
