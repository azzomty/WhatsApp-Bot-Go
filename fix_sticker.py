with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    c = f.read()

# Fix non-bulk media parsing
old_media = '''} else if ext := unwrappedMsg.GetExtendedTextMessage(); ext != nil {
			quoted := UnwrapMessage(ext.GetContextInfo().GetQuotedMessage())
			if qImg := quoted.GetImageMessage(); qImg != nil {
				mediaMsg = qImg
			} else if qVid := quoted.GetVideoMessage(); qVid != nil {
				mediaMsg = qVid
				isVideo = true
			} else if qDoc := quoted.GetDocumentMessage(); qDoc != nil && (strings.HasPrefix(qDoc.GetMimetype(), "image/") || strings.HasPrefix(qDoc.GetMimetype(), "video/")) {
				mediaMsg = qDoc
				isVideo = strings.HasPrefix(qDoc.GetMimetype(), "video/")
			}
		}
		if mediaMsg == nil {
			sendMessage(ctx, "أرسل صورة أو فيديو مع الأمر، أو رد على صورة/فيديو.")
			return
		}'''

new_media = '''} else if ext := unwrappedMsg.GetExtendedTextMessage(); ext != nil {
			quoted := UnwrapMessage(ext.GetContextInfo().GetQuotedMessage())
			if qImg := quoted.GetImageMessage(); qImg != nil {
				mediaMsg = qImg
			} else if qVid := quoted.GetVideoMessage(); qVid != nil {
				mediaMsg = qVid
				isVideo = true
			} else if qDoc := quoted.GetDocumentMessage(); qDoc != nil && (strings.HasPrefix(qDoc.GetMimetype(), "image/") || strings.HasPrefix(qDoc.GetMimetype(), "video/")) {
				mediaMsg = qDoc
				isVideo = strings.HasPrefix(qDoc.GetMimetype(), "video/")
			} else if qSticker := quoted.GetStickerMessage(); qSticker != nil {
				mediaMsg = qSticker
				isVideo = qSticker.GetIsAnimated()
			}
		}
		if mediaMsg == nil {
			sendMessage(ctx, "أرسل صورة أو فيديو مع الأمر، أو رد على صورة/فيديو/ملصق.")
			return
		}'''

c = c.replace(old_media, new_media)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(c)
