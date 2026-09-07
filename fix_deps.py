with open('/home/lennox/Desktop/اهها/Go_Bot/download_deps.go', 'r') as f:
    content = f.read()

webpmux_code = """
	// Download webpmux
	info3, err3 := os.Stat("webpmux")
	if os.IsNotExist(err3) || (err3 == nil && info3.Size() < 100000) {
		fmt.Println("Downloading webpmux...")
		cmd := exec.Command("curl", "-L", "-o", "libwebp.tar.gz", "https://storage.googleapis.com/downloads.webmproject.org/releases/webp/libwebp-1.3.2-linux-x86-64.tar.gz")
		cmd.Run()
		exec.Command("tar", "-xzf", "libwebp.tar.gz").Run()
		exec.Command("cp", "libwebp-1.3.2-linux-x86-64/bin/webpmux", ".").Run()
		os.Remove("libwebp.tar.gz")
		os.RemoveAll("libwebp-1.3.2-linux-x86-64")
		os.Chmod("webpmux", 0755)
		fmt.Println("webpmux downloaded.")
	}
}"""

content = content.replace('}\n\n', '}\n')
# Just append it at the end of initDeps()
if 'webpmux' not in content:
    content = content.replace('}\n', webpmux_code, 1) # wait, replacing the last brace

with open('/home/lennox/Desktop/اهها/Go_Bot/download_deps.go', 'w') as f:
    f.write(content)
print("Patched deps")
