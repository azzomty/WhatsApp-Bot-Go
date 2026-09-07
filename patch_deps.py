with open('/home/lennox/Desktop/اهها/Go_Bot/download_deps.go', 'r') as f:
    content = f.read()

new_deps = """
	// Download webpmux for sticker EXIF
	if _, err := os.Stat("webpmux"); os.IsNotExist(err) {
		fmt.Println("Downloading webpmux...")
		webpURL := "https://storage.googleapis.com/downloads.webmproject.org/releases/webp/libwebp-1.3.2-linux-x86-64.tar.gz"
		cmd := exec.Command("curl", "-L", "-o", "libwebp.tar.gz", webpURL)
		cmd.Run()
		
		exec.Command("tar", "-xzf", "libwebp.tar.gz").Run()
		exec.Command("cp", "libwebp-1.3.2-linux-x86-64/bin/webpmux", ".").Run()
		
		os.Remove("libwebp.tar.gz")
		os.RemoveAll("libwebp-1.3.2-linux-x86-64")
	}
"""

content = content.replace('fmt.Println("Dependencies initialized successfully!")', new_deps + '\n\tfmt.Println("Dependencies initialized successfully!")')

with open('/home/lennox/Desktop/اهها/Go_Bot/download_deps.go', 'w') as f:
    f.write(content)
print("Patched deps")
