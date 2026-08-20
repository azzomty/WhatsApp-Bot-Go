package commands

import (
	"github.com/bregydoc/gtranslate"
	"strings"
)

func translateMessage(ctx *BotContext, targetLang string) {
	if ctx.Event.Message.GetExtendedTextMessage() == nil || ctx.Event.Message.GetExtendedTextMessage().GetContextInfo() == nil || ctx.Event.Message.GetExtendedTextMessage().GetContextInfo().GetQuotedMessage() == nil {
		sendMessage(ctx, "يرجى الرد (Reply) على رسالة نصية واستخدام الأمر")
		return
	}

	quoted := ctx.Event.Message.GetExtendedTextMessage().GetContextInfo().GetQuotedMessage()
	textToTranslate := ""

	if quoted.GetConversation() != "" {
		textToTranslate = quoted.GetConversation()
	} else if quoted.GetExtendedTextMessage() != nil {
		textToTranslate = quoted.GetExtendedTextMessage().GetText()
	} else if quoted.GetImageMessage() != nil {
		textToTranslate = quoted.GetImageMessage().GetCaption()
	} else if quoted.GetVideoMessage() != nil {
		textToTranslate = quoted.GetVideoMessage().GetCaption()
	}

	textToTranslate = strings.TrimSpace(textToTranslate)
	if textToTranslate == "" {
		sendMessage(ctx, "الرسالة لا تحتوي على نص قابل للترجمة.")
		return
	}

	translated, err := gtranslate.TranslateWithParams(
		textToTranslate,
		gtranslate.TranslationParams{
			From: "auto",
			To:   targetLang,
		},
	)

	if err != nil {
		sendMessage(ctx, "حدث خطأ أثناء الترجمة ")
		return
	}

	sendMessage(ctx, translated)
}
