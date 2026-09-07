import re

with open("internal/pinterest/pinterest.go", "r") as f:
    content = f.read()

target_pending = """type PendingRequest struct {
	Query       string
	Count       int
	IsVisual    bool
	Base64Image string
}

func SetPending(chatID, query string, count int, isVisual bool, base64Image string) {
	pendingMutex.Lock()
	defer pendingMutex.Unlock()
	PendingRequests[chatID] = PendingRequest{Query: query, Count: count, IsVisual: isVisual, Base64Image: base64Image}
}"""

new_pending = """type PendingRequest struct {
	Query       string
	Count       int
	IsVisual    bool
	Base64Image string
	Bookmark    string
}

func SetPending(chatID, query string, count int, isVisual bool, base64Image string, bookmark string) {
	pendingMutex.Lock()
	defer pendingMutex.Unlock()
	PendingRequests[chatID] = PendingRequest{Query: query, Count: count, IsVisual: isVisual, Base64Image: base64Image, Bookmark: bookmark}
}"""
content = content.replace(target_pending, new_pending)

with open("internal/pinterest/pinterest.go", "w") as f:
    f.write(content)
