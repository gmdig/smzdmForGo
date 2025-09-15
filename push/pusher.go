package push

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"ggball.com/smzdm/file"
	"ggball.com/smzdm/smzdm"
)


// 内部方法：发送消息到 Telegram
func sendTelegramMessage(conf file.Config, text string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", conf.TelegramBotToken)

	body := map[string]interface{}{
		"chat_id":    conf.TelegramChatID,
		"text":       text,
		"parse_mode": "Markdown",
	}

	data, _ := json.Marshal(body)
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram push failed, status: %s", resp.Status)
	}
	return nil
}

// --- 以下函数替换原来的 DingTalk 版本 ---

// 推送商品（列表）
func PushProWithDingDing(pro []smzdm.Product, conf file.Config) {
	if len(pro) == 0 {
		return
	}

	msg := "【好物推荐】\n"
	for i, item := range pro {
		if i >= conf.SatisfyNum {
			break
		}
		msg += fmt.Sprintf("[%s](%s) - %s\n", item.ArticleTitle, item.ArticleUrl, item.ArticlePrice)
	}

	_ = sendTelegramMessage(conf, msg)
}

// 推送纯文本
func PushTextWithDingDing(resText string, conf file.Config) {
	msg := resText + " 【什么值得买】"
	_ = sendTelegramMessage(conf, msg)
}

// 推送文字并 @用户（Telegram 没有手机号 @，只能直接发群消息）
func PushTextWithDingDingWIthMoblie(pro []smzdm.Product, conf file.Config, atMobiles []string) {
	if len(pro) == 0 {
		return
	}

	msg := "【好物到了】\n"
	for _, item := range pro {
		msg += fmt.Sprintf("[%s](%s) - %s\n", item.ArticleTitle, item.ArticleUrl, item.ArticlePrice)
	}

	// Telegram 无法按手机号 @人，这里直接推送文本
	_ = sendTelegramMessage(conf, msg)
}
