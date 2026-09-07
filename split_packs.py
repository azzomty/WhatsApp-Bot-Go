import re

with open("internal/commands/sticker_pack.go", "r") as f:
    c = f.read()

target = """	for i, img := range images {
		w, _ = zipWriter.Create(fmt.Sprintf("%d.webp", i+1))
		w.Write(img)
	}
	zipWriter.Close()

	uploaded, err := ctx.Client.Upload(context.Background(), buf.Bytes(), whatsmeow.MediaDocument)
	if err != nil {
		sendMessage(ctx, "فشل رفع الحزمة!")
		return
	}

	fileName := session.Title + ".wastickers"
	msg := &waProto.Message{
		DocumentMessage: &waProto.DocumentMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String("application/zip"),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(buf.Len())),
			FileName:      proto.String(fileName),
		},
	}
	ctx.Client.SendMessage(context.Background(), ctx.Event.Info.Chat, msg)"""

new_target = """
	// We need to split into multiple packs if > 30 stickers because WhatsApp has a hard limit of 30 per pack
	numPacks := (len(images) + 29) / 30

	for packIdx := 0; packIdx < numPacks; packIdx++ {
		startIdx := packIdx * 30
		endIdx := startIdx + 30
		if endIdx > len(images) {
			endIdx = len(images)
		}
		
		packImages := images[startIdx:endIdx]
		
		buf := new(bytes.Buffer)
		zipWriter := zip.NewWriter(buf)
		
		packTitle := session.Title
		if numPacks > 1 {
			packTitle = fmt.Sprintf("%s Part %d", session.Title, packIdx+1)
		}

		w, _ := zipWriter.Create("title.txt")
		w.Write([]byte(packTitle))

		w, _ = zipWriter.Create("author.txt")
		w.Write([]byte(session.Author))

		wTray, _ := zipWriter.Create("tray.png")
		if len(packImages) > 0 {
			trayResp, err := http.Post("http://127.0.0.1:4321/tray", "application/octet-stream", bytes.NewReader(packImages[0]))
			if err == nil && trayResp.StatusCode == 200 {
				trayImg, _ := ioutil.ReadAll(trayResp.Body)
				wTray.Write(trayImg)
			} else {
				imgTray := image.NewRGBA(image.Rect(0, 0, 96, 96))
				png.Encode(wTray, imgTray)
			}
		}

		for i, img := range packImages {
			w, _ = zipWriter.Create(fmt.Sprintf("%d.webp", i+1))
			w.Write(img)
		}
		zipWriter.Close()

		uploaded, err := ctx.Client.Upload(context.Background(), buf.Bytes(), whatsmeow.MediaDocument)
		if err != nil {
			sendMessage(ctx, fmt.Sprintf("فشل رفع الحزمة %d!", packIdx+1))
			continue
		}

		fileName := packTitle + ".wastickers"
		msg := &waProto.Message{
			DocumentMessage: &waProto.DocumentMessage{
				URL:           proto.String(uploaded.URL),
				DirectPath:    proto.String(uploaded.DirectPath),
				MediaKey:      uploaded.MediaKey,
				Mimetype:      proto.String("application/zip"),
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    proto.Uint64(uint64(buf.Len())),
				FileName:      proto.String(fileName),
			},
		}
		ctx.Client.SendMessage(context.Background(), ctx.Event.Info.Chat, msg)
	}
"""
c = c.replace(target, new_target)

# We also need to remove the top-level tray generation that is now duplicated!
target_top_tray = """	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	w, _ := zipWriter.Create("title.txt")
	w.Write([]byte(session.Title))

	w, _ = zipWriter.Create("author.txt")
	w.Write([]byte(session.Author))

	// Create a transparent 96x96 tray image with a visible icon so WhatsApp doesn't reject it
	// Generate a real 96x96 tray image from the first sticker
	wTray, _ := zipWriter.Create("tray.png")
	if len(images) > 0 {
		trayResp, err := http.Post("http://127.0.0.1:4321/tray", "application/octet-stream", bytes.NewReader(images[0]))
		if err == nil && trayResp.StatusCode == 200 {
			trayImg, _ := ioutil.ReadAll(trayResp.Body)
			wTray.Write(trayImg)
		} else {
			// Fallback to empty if it fails
			imgTray := image.NewRGBA(image.Rect(0, 0, 96, 96))
			png.Encode(wTray, imgTray)
		}
	}"""
c = c.replace(target_top_tray, "")

with open("internal/commands/sticker_pack.go", "w") as f:
    f.write(c)

