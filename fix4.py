with open("internal/pinterest/pinterest.go", "r") as f:
    p = f.read()
p = p.replace('''type PendingRequest struct {
	Query       string
	Count       int
	IsVisual    bool
	Base64Image string
}''', '''type PendingRequest struct {
	Query       string
	Count       int
	IsVisual    bool
	Base64Image string
	Bookmark    string
}''')
with open("internal/pinterest/pinterest.go", "w") as f:
    f.write(p)
