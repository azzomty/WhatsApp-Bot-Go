import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'r') as f:
    content = f.read()

# Replace the > 64MB split logic with just sending the video normally.
# Look for sendVideoData in media.go

old_logic = """	if len(data) <= 64*1024*1024 {
		// Just send it normally
		resp, err := ctx.Client.Upload(context.Background(), data, whatsmeow.MediaVideo)
		if err != nil {
			sendMessage(ctx, "فشل في رفع الحلقة للواتساب.")
			return
		}

		vidMsg := &waProto.VideoMessage{
			URL:           proto.String(resp.URL),
			DirectPath:    proto.String(resp.DirectPath),
			MediaKey:      resp.MediaKey,
			Mimetype:      proto.String("video/mp4"),
			FileEncSHA256: resp.FileEncSHA256,
			FileSHA256:    resp.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
			Caption:       proto.String(fmt.Sprintf("*%s* - الحلقة %s", animeName, epNum)),
		}

		ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
			VideoMessage: vidMsg,
		})
		return
	}
	
	// File is > 64MB, split it!
	sendMessage(ctx, "حجم الحلقة كبير جداً (أكثر من 64 ميجا)، جاري تقسيمها إلى أجزاء وإرسالها... ")
	
	tempDir, err := os.MkdirTemp("", "video_split")
	if err != nil {
		sendMessage(ctx, "فشل في إنشاء مجلد مؤقت للتقسيم.")
		return
	}
	defer os.RemoveAll(tempDir)
	
	inputPath := tempDir + "/input.mp4"
	if err := os.WriteFile(inputPath, data, 0644); err != nil {
		sendMessage(ctx, "فشل في كتابة الملف المؤقت.")
		return
	}
	
	outPattern := tempDir + "/part_%03d.mp4"
	
	ffmpegPath := "node_modules/ffmpeg-static/ffmpeg"
	if runtime.GOOS == "windows" {
		if _, err := os.Stat("node_modules/ffmpeg-static/ffmpeg.exe"); err == nil {
			ffmpegPath = "node_modules/ffmpeg-static/ffmpeg.exe"
		}
	}
	
	cmd := exec.Command(ffmpegPath, "-i", inputPath, "-c", "copy", "-f", "segment", "-segment_time", "600", "-reset_timestamps", "1", outPattern)
	if err := cmd.Run(); err != nil {
		sendMessage(ctx, "فشل في تقسيم الفيديو. قد تكون الحلقة بصيغة غير مدعومة للتقسيم السريع.")
		return
	}
	
	files, err := os.ReadDir(tempDir)
	if err != nil {
		return
	}
	
	var parts []string
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "part_") {
			parts = append(parts, tempDir+"/"+f.Name())
		}
	}
	
	// Send each part
	for i, partPath := range parts {
		partData, err := os.ReadFile(partPath)
		if err != nil {
			continue
		}
		
		resp, err := ctx.Client.Upload(context.Background(), partData, whatsmeow.MediaVideo)
		if err != nil {
			continue
		}
		
		caption := fmt.Sprintf("*%s* - الحلقة %s\\n(الجزء %d من %d)", animeName, epNum, i+1, len(parts))
		
		vidMsg := &waProto.VideoMessage{
			URL:           proto.String(resp.URL),
			DirectPath:    proto.String(resp.DirectPath),
			MediaKey:      resp.MediaKey,
			Mimetype:      proto.String("video/mp4"),
			FileEncSHA256: resp.FileEncSHA256,
			FileSHA256:    resp.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(partData))),
			Caption:       proto.String(caption),
		}
		
		ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
			VideoMessage: vidMsg,
		})
	}"""

new_logic = """	// Just send it normally without splitting
	resp, err := ctx.Client.Upload(context.Background(), data, whatsmeow.MediaVideo)
	if err != nil {
		sendMessage(ctx, "فشل في رفع الحلقة للواتساب.")
		return
	}

	vidMsg := &waProto.VideoMessage{
		URL:           proto.String(resp.URL),
		DirectPath:    proto.String(resp.DirectPath),
		MediaKey:      resp.MediaKey,
		Mimetype:      proto.String("video/mp4"),
		FileEncSHA256: resp.FileEncSHA256,
		FileSHA256:    resp.FileSHA256,
		FileLength:    proto.Uint64(uint64(len(data))),
		Caption:       proto.String(fmt.Sprintf("*%s* - الحلقة %s", animeName, epNum)),
	}

	ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
		VideoMessage: vidMsg,
	})"""

content = content.replace(old_logic, new_logic)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'w') as f:
    f.write(content)
