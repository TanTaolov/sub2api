package service

import "strings"

// AccountTestModeCandy 是 OpenAI 账号连通性测试的「糖果测试」模式：不发常规的
// 极短探活提示词，而是发一道固定的抽屉原理（鸽巢原理）推理题，用于观察账号背后
// 模型的真实推理与输出质量。题面固定不可改，便于跨账号对比同一道题的回答。
const AccountTestModeCandy = "candy"

// openAICandyTestPrompt 是糖果测试的固定题面。
const openAICandyTestPrompt = `在一个黑色的袋子里放有三种口味的糖果，每种糖果有两种不同的形状（圆形和五角星形，不同的形状靠手感可以分辨）。现已知不同口味的糖果和不同形状的数量统计如下表：

| 形状 | 苹果味 | 桃子味 | 西瓜味 |
| --- | --- | --- | --- |
| 圆形 | 7 | 9 | 8 |
| 五角星形 | 7 | 6 | 4 |

参赛者需要在活动前决定摸出的糖果数目，那么，最少取出多少个糖果才能保证手中同时拥有不同形状的苹果味和桃子味的糖果？（同时手中有圆形苹果味匹配五角星桃子味糖果，或者有圆形桃子味匹配五角星苹果味糖果都满足要求）`

// resolveOpenAITextTestPrompt 决定文本类 OpenAI 测试请求实际发送的 user 内容：
// 糖果测试始终发送固定题面（忽略调用方 prompt），其余模式沿用调用方 prompt，
// 为空时回退到默认极短探活提示词。
func resolveOpenAITextTestPrompt(mode string, prompt string) string {
	if mode == AccountTestModeCandy {
		return openAICandyTestPrompt
	}
	if trimmed := strings.TrimSpace(prompt); trimmed != "" {
		return trimmed
	}
	return defaultOpenAITextTestPrompt
}
