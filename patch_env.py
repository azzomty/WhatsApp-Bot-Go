import re

with open("main.go", "r") as f:
    c = f.read()

# Add import
target_import = """	"time"

	"go.mau.fi/whatsmeow" """
new_import = """	"time"
	"github.com/joho/godotenv"
	"go.mau.fi/whatsmeow" """
c = c.replace(target_import, new_import)

# Add load
target_main = """func main() {
	// Initialize store"""
new_main = """func main() {
	godotenv.Load()
	// Initialize store"""
c = c.replace(target_main, new_main)

with open("main.go", "w") as f:
    f.write(c)

