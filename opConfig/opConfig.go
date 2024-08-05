package opConfig

import "strings"

var UserConf = map[string][]string{
	"fox": {"tse_2330.tw", "tse_0050.tw"},
}

func GetUserStocks(user string) string {
	stock, ok := UserConf[user]

	if !ok {
		return ""
	}

	return strings.Join(stock, "|")
}
