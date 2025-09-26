package smzdm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
	"regexp"

	"ggball.com/smzdm/file"
)

type result struct {
	ErrorCode string `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
	Data      Data   `json:"data"`
}

type Data struct {
	Rows  []Product `json:"rows"`
	Total int       `json:"total"`
}

type Product struct {
	ArticleTitle   string `json:"article_title"`
	ArticlePrice   string `json:"article_price"`
	ArticleWorthy  string `json:"article_worthy"`
	ArticleComment string `json:"article_comment"`
	ArticleId      string `json:"article_id"`
	ArticleDate    string `json:"publish_date_lt"`
	ArticlePic     string `json:"article_pic"`
	ArticleUrl     string `json:"article_url"`
	Referral       string `json:"article_referrals"`
}

// 全局配置
var globalConf = file.Config{}

// 推送信息文件地址
var pushedPath = "./pushed.json"

// 获取商品
func GetSatisfiedGoods(conf file.Config) ([]Product, []Product) {
	globalConf = conf
	fmt.Println("开始爬取符合条件商品。。")

	// 获取已推送文章id
	pushedMap := file.ReadPusedInfo(pushedPath)

	// 符合条件的商品集合
	var satisfyGoodsList []Product

	page := 0
	for {
		productList := GetGoods(page, "").Data.Rows

		if len(productList) > 0 {
			for _, good := range productList {
				// 评论数包含“K”时，默认给 1000
				if strings.Contains(strings.ToLower(good.ArticleComment), "k") {
					good.ArticleComment = "1000"
				}

				if removeByFilterRules(good, pushedMap) {
					continue
				}

				if satisfy(good, satisfyGoodsList) {
					satisfyGoodsList = append(satisfyGoodsList, good)
				}
			}
		}

		page++
		time.Sleep(time.Duration(2) * time.Second)

		if shouldStop(len(satisfyGoodsList), page) {
			fmt.Println("退出")
			break
		}
	}

	// 评论数排序
	sort.SliceStable(satisfyGoodsList, func(a, b int) bool {
		return strings.Compare(satisfyGoodsList[a].ArticleComment, satisfyGoodsList[b].ArticleComment) > 0
	})

	fmt.Println("结束爬取符合条件商品。。")

	// 自己的商品
	satisfyGoodsListBySelf := filterMyselfProduct(satisfyGoodsList)

	// 保存推送商品
	savePushed(pushedMap, pushedPath, satisfyGoodsList)

	return satisfyGoodsList, satisfyGoodsListBySelf
}

// 获取商品集合
func GetGoods(page int, keword string) result {
	var res result

	params := url.Values{}
	Url, err := url.Parse("https://api.smzdm.com/v1/list")
	if err != nil {
		return res
	}
	params.Set("keyword", keword)
	params.Set("order", "time")
	params.Set("type", "good_price")
	params.Set("offset", strconv.Itoa(page*100))
	params.Set("limit", "100")

	Url.RawQuery = params.Encode()
	urlPath := Url.String()
	fmt.Println(urlPath)
	resp, err := http.Get(urlPath)
	if err != nil {
		return res
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	_ = json.Unmarshal(body, &res)
	return res
}

// 根据条件 判断是否应该停止爬取
func shouldStop(length int, page int) bool {
	fmt.Println("length:" + strconv.Itoa(length) + "\n\r page:" + strconv.Itoa(page))
	return length > globalConf.SatisfyNum || page > 100
}

// 根据过滤规则，去除商品
func removeByFilterRules(good Product, pushedMap map[string]interface{}) bool {
	var noNeed = false

	// 1. 标题/价格包含过滤词
	for _, word := range globalConf.FilterWords {
		var pattern string
		if strings.HasPrefix(word, "re:") {
			pattern = "(?i)" + word[3:]
		} else {
			pattern = "(?i)" + regexp.QuoteMeta(word)
		}

		if matched, _ := regexp.MatchString(pattern, good.ArticleTitle); matched {
			fmt.Printf("过滤掉(标题): %s by %s\n", good.ArticleTitle, word)
			noNeed = true
			break
		}
		if matched, _ := regexp.MatchString(pattern, good.ArticlePrice); matched {
			fmt.Printf("过滤掉(价格): %s by %s\n", good.ArticlePrice, word)
			noNeed = true
			break
		}
	}

	// 2. 已推送过
	if _, exists := pushedMap[good.ArticleId]; exists {
		noNeed = true
	}

	// 3. 时间小于前天
	nTime := time.Now()
	beforeYesDate := nTime.AddDate(0, 0, -2)
	dateInt64, err1 := strconv.ParseInt(good.ArticleDate, 10, 64)
	if err1 != nil {
		panic(err1)
	}
	arDate := time.Unix(dateInt64, 0)
	if arDate.Before(beforeYesDate) {
		noNeed = true
	}

	return noNeed
}

// 根据规则判断符合规则的商品
func satisfy(good Product, satisfyGoodsList []Product) bool {
	articleComment, err1 := strconv.Atoi(good.ArticleComment)
	articleWorthy, err2 := strconv.Atoi(good.ArticleWorthy)

	if err1 != nil || err2 != nil {
		fmt.Println("goods:", good)
		panic(err1)
	}

	if articleComment >= globalConf.LowCommentNum || articleWorthy >= globalConf.LowWorthyNum {
		fmt.Printf("appear satisfy good: %#v", good)
		return true
	}
	return false
}

// 保存推送商品
func savePushed(pushedMap map[string]interface{}, pushedPath string, satisfyGoodsList []Product) {
	tempMap := make(map[string]interface{})
	for index, value := range satisfyGoodsList {
		tempMap[value.ArticleId] = index
	}
	file.WritePushedInfo(tempMap, pushedMap, pushedPath)
}

// 过滤自己的商品
func filterMyselfProduct(satisfyGoodsList []Product) []Product {
	var satisfyGoodsListBySelf []Product

	for _, value := range satisfyGoodsList {
		for _, word := range globalConf.KeyWords {
			var pattern string
			if strings.HasPrefix(word, "re:") {
				pattern = "(?i)" + word[3:]
			} else {
				pattern = "(?i)" + regexp.QuoteMeta(word)
			}

			matched, err := regexp.MatchString(pattern, value.ArticleTitle)
			if err != nil {
				fmt.Printf("正则表达式匹配错误: %v (word=%s)\n", err, word)
				continue
			}
			if matched {
				fmt.Printf("appear myself satisfy good: %#v\n", value)
				satisfyGoodsListBySelf = append(satisfyGoodsListBySelf, value)
				break
			}
		}
	}
	return satisfyGoodsListBySelf
}
