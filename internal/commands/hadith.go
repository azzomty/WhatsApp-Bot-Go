package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

type Category struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Count string `json:"hadeeths_count"`
}

var (
	allCategories []Category
	categoriesMap = make(map[string]Category)
)

func init() {
	rand.Seed(time.Now().UnixNano())
	go fetchAllCategories()
}

func fetchAllCategories() {
	resp, err := http.Get("https://hadeethenc.com/api/v1/categories/list/?language=ar")
	if err == nil {
		defer resp.Body.Close()
		json.NewDecoder(resp.Body).Decode(&allCategories)
		for _, c := range allCategories {
			categoriesMap[c.ID] = c
		}
	}
}

func HandleHadithMenu(ctx *BotContext) {
	if len(allCategories) == 0 {
		sendMessage(ctx, "جاري جلب الأقسام من السيرفر، انتظر ثواني وجرب مرة ثانية...")
		return
	}

	msg := "*قائمة أقسام الأحاديث*\n\nالعدد الكلي للأقسام: " + strconv.Itoa(len(allCategories)) + "\n\n"
	msg += "بسبب كثرة الأقسام، يمكنك البحث عن قسم معين بكتابة:\n"
	msg += "`.بحث قسم <كلمة>`\n\n"
	msg += "مثال:\n"
	msg += "`.بحث قسم البخاري`\n"
	msg += "`.بحث قسم الصلاة`\n\n"
	msg += "ولأخذ حديث عشوائي من قسم معين، اكتب:\n"
	msg += "`.حديث <رقم القسم>`\n"
	
	ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{Text: proto.String(msg)},
	})
}

func HandleAllCategories(ctx *BotContext) {
	if len(allCategories) == 0 {
		sendMessage(ctx, "جاري جلب الأقسام من السيرفر، انتظر ثواني وجرب مرة ثانية...")
		return
	}
	
	msg := fmt.Sprintf("*جميع أقسام الأحاديث (%d قسم)*\n\n", len(allCategories))
	var lines []string
	for i, c := range allCategories {
		lines = append(lines, fmt.Sprintf("*%d.* %s (عدد الأحاديث: %s)", i+1, c.Title, c.Count))
	}
	msg += strings.Join(lines, "\n")
	msg += "\n\nلاستخراج حديث، اكتب `.حديث <رقم القسم>`"

	ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{Text: proto.String(msg)},
	})
}

func HandleCategorySearch(ctx *BotContext) {
	if len(allCategories) == 0 {
		sendMessage(ctx, "جاري جلب الأقسام من السيرفر، انتظر ثواني وجرب مرة ثانية...")
		return
	}

	query := strings.TrimPrefix(ctx.Text, ".بحث قسم ")
	if query == "" || query == ctx.Text {
		sendMessage(ctx, "اكتب الكلمة اللي تبي تبحث عنها، مثال: .بحث قسم الصيام")
		return
	}

	results := []string{}
	for i, c := range allCategories {
		if strings.Contains(c.Title, query) {
			results = append(results, fmt.Sprintf("*%d.* %s (عدد الأحاديث: %s)", i+1, c.Title, c.Count))
		}
	}

	if len(results) == 0 {
		sendMessage(ctx, "لم يتم العثور على أي قسم يطابق بحثك.")
		return
	}

	msg := "*نتائج البحث:*\n\n"
	if len(results) > 50 {
		results = results[:50]
		msg += "(تم إظهار أول 50 نتيجة فقط)\n\n"
	}
	msg += strings.Join(results, "\n")
	msg += "\n\nلاستخراج حديث، اكتب `.حديث <رقم القسم>`"

	ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{Text: proto.String(msg)},
	})
}

type HadeethListResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

type HadeethOneResponse struct {
	Hadeeth     string `json:"hadeeth"`
	Explanation string `json:"explanation"`
}

func HandleHadith(ctx *BotContext) {
	if len(allCategories) == 0 {
		sendMessage(ctx, "السيرفر يجهز الأقسام، ثواني وجرب...")
		return
	}

	categoryID := ""
	parts := strings.Split(ctx.Text, " ")
	if len(parts) > 1 {
		idx, err := strconv.Atoi(parts[1])
		if err == nil && idx >= 1 && idx <= len(allCategories) {
			categoryID = allCategories[idx-1].ID
		}
	}

	if categoryID == "" {
		// Pick a random category that has > 0 hadiths
		valid := []Category{}
		for _, c := range allCategories {
			count, _ := strconv.Atoi(c.Count)
			if count > 0 {
				valid = append(valid, c)
			}
		}
		if len(valid) > 0 {
			categoryID = valid[rand.Intn(len(valid))].ID
		}
	}

	catName := categoriesMap[categoryID].Title
	countStr := categoriesMap[categoryID].Count
	count, _ := strconv.Atoi(countStr)
	
	// Randomize the page! This fixes the issue of always getting the same hadith in large categories
	page := 1
	if count > 100 {
		maxPages := count / 100
		if count%100 != 0 {
			maxPages++
		}
		page = rand.Intn(maxPages) + 1
	}

	// 1. Get list of hadiths in this category
	listURL := fmt.Sprintf("https://hadeethenc.com/api/v1/hadeeths/list/?language=ar&category_id=%s&page=%d&per_page=100", categoryID, page)
	resp, err := http.Get(listURL)
	if err != nil {
		sendMessage(ctx, "حدث خطأ أثناء جلب الحديث. جرب مرة ثانية.")
		return
	}
	defer resp.Body.Close()

	var listData HadeethListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listData); err != nil || len(listData.Data) == 0 {
		sendMessage(ctx, "حدث خطأ في قراءة بيانات الحديث. قد يكون القسم فارغاً.")
		return
	}

	// 2. Pick a random hadith ID from the page
	randomItem := listData.Data[rand.Intn(len(listData.Data))]

	// 3. Get the hadith text
	oneURL := fmt.Sprintf("https://hadeethenc.com/api/v1/hadeeths/one/?language=ar&id=%s", randomItem.ID)
	resp2, err := http.Get(oneURL)
	if err != nil {
		sendMessage(ctx, "حدث خطأ أثناء جلب نص الحديث.")
		return
	}
	defer resp2.Body.Close()

	var oneData HadeethOneResponse
	if err := json.NewDecoder(resp2.Body).Decode(&oneData); err != nil {
		sendMessage(ctx, "حدث خطأ في قراءة نص الحديث.")
		return
	}

	msg := fmt.Sprintf("*حديث نبوي* (قسم: %s)\n\n", catName)
	msg += fmt.Sprintf("« %s »\n\n", strings.TrimSpace(oneData.Hadeeth))
	msg += fmt.Sprintf("*الشرح:*\n%s", strings.TrimSpace(oneData.Explanation))

	ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{Text: proto.String(msg)},
	})
}
