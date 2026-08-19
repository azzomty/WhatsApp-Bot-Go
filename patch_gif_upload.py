import re

with open("main.go", "r") as f:
    content = f.read()

target = """							mediaType := whatsmeow.MediaImage
							if isVid {
								mediaType = whatsmeow.MediaVideo
							}

							resp, err := client.Upload(context.Background(), data, mediaType)
							if err == nil {
								msg := &waProto.Message{}
								if isVid {
									vidMsg := &waProto.VideoMessage{
										URL:           proto.String(resp.URL),
										DirectPath:    proto.String(resp.DirectPath),
										MediaKey:      resp.MediaKey,
										Mimetype:      proto.String("video/mp4"),
										FileEncSHA256: resp.FileEncSHA256,
										FileSHA256:    resp.FileSHA256,
										FileLength:    proto.Uint64(uint64(len(data))),
										ContextInfo: &waProto.ContextInfo{
											StanzaID:      proto.String(v.Info.ID),
											Participant:   proto.String(v.Info.Sender.String()),
											QuotedMessage: v.Message,
										},
									}
									if aspect == "gif" {
										vidMsg.GifPlayback = proto.Bool(true)
									}
									msg.VideoMessage = vidMsg
								} else {
									imgMsg := &waProto.ImageMessage{
										URL:           proto.String(resp.URL),
										DirectPath:    proto.String(resp.DirectPath),
										MediaKey:      resp.MediaKey,
										Mimetype:      proto.String("image/jpeg"),
										FileEncSHA256: resp.FileEncSHA256,
										FileSHA256:    resp.FileSHA256,
										FileLength:    proto.Uint64(uint64(len(data))),
										ContextInfo: &waProto.ContextInfo{
											StanzaID:      proto.String(v.Info.ID),
											Participant:   proto.String(v.Info.Sender.String()),
											QuotedMessage: v.Message,
										},
									}
									msg.ImageMessage = imgMsg
								}

								client.SendMessage(context.Background(), v.Info.Chat, msg)"""

new_code = """							mediaType := whatsmeow.MediaImage
							if isVid {
								mediaType = whatsmeow.MediaVideo
							} else if aspect == "gif" {
								mediaType = whatsmeow.MediaDocument
							}

							resp, err := client.Upload(context.Background(), data, mediaType)
							if err == nil {
								msg := &waProto.Message{}
								if isVid {
									vidMsg := &waProto.VideoMessage{
										URL:           proto.String(resp.URL),
										DirectPath:    proto.String(resp.DirectPath),
										MediaKey:      resp.MediaKey,
										Mimetype:      proto.String("video/mp4"),
										FileEncSHA256: resp.FileEncSHA256,
										FileSHA256:    resp.FileSHA256,
										FileLength:    proto.Uint64(uint64(len(data))),
										ContextInfo: &waProto.ContextInfo{
											StanzaID:      proto.String(v.Info.ID),
											Participant:   proto.String(v.Info.Sender.String()),
											QuotedMessage: v.Message,
										},
									}
									if aspect == "gif" {
										vidMsg.GifPlayback = proto.Bool(true)
									}
									msg.VideoMessage = vidMsg
								} else if aspect == "gif" {
									docMsg := &waProto.DocumentMessage{
										URL:           proto.String(resp.URL),
										DirectPath:    proto.String(resp.DirectPath),
										MediaKey:      resp.MediaKey,
										Mimetype:      proto.String("image/gif"),
										FileEncSHA256: resp.FileEncSHA256,
										FileSHA256:    resp.FileSHA256,
										FileLength:    proto.Uint64(uint64(len(data))),
										FileName:      proto.String("animated.gif"),
										ContextInfo: &waProto.ContextInfo{
											StanzaID:      proto.String(v.Info.ID),
											Participant:   proto.String(v.Info.Sender.String()),
											QuotedMessage: v.Message,
										},
									}
									msg.DocumentMessage = docMsg
								} else {
									imgMsg := &waProto.ImageMessage{
										URL:           proto.String(resp.URL),
										DirectPath:    proto.String(resp.DirectPath),
										MediaKey:      resp.MediaKey,
										Mimetype:      proto.String("image/jpeg"),
										FileEncSHA256: resp.FileEncSHA256,
										FileSHA256:    resp.FileSHA256,
										FileLength:    proto.Uint64(uint64(len(data))),
										ContextInfo: &waProto.ContextInfo{
											StanzaID:      proto.String(v.Info.ID),
											Participant:   proto.String(v.Info.Sender.String()),
											QuotedMessage: v.Message,
										},
									}
									msg.ImageMessage = imgMsg
								}

								client.SendMessage(context.Background(), v.Info.Chat, msg)"""

content = content.replace(target, new_code)

with open("main.go", "w") as f:
    f.write(content)

