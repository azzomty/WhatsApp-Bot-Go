import re
import requests

html = requests.get('https://uqload.vc/embed-0a0tvjrbcfoi.html').text

match = re.search(r'eval\(function\(p,a,c,k,e,d\).*?return p}\(\'(.*?)\',(\d+),(\d+),\'(.*?)\'\.split', html)
if not match:
    print("Could not find eval packer")
    exit(1)

p = match.group(1)
a = int(match.group(2))
c = int(match.group(3))
k = match.group(4).split('|')

def baseN(num, b):
    if num == 0:
        return "0"
    res = ""
    while num > 0:
        remainder = num % b
        if remainder > 35:
            res = chr(remainder + 29) + res
        elif remainder > 9:
            res = chr(remainder + 87) + res
        else:
            res = str(remainder) + res
        num //= b
    return res

for i in range(c - 1, -1, -1):
    if k[i]:
        p = re.sub(r'\b' + baseN(i, a) + r'\b', k[i], p)

print(p[:500])
import json
m3u8_match = re.search(r'file\s*:\s*["\'](https?://[^"\']+\.m3u8[^"\']*)["\']', p)
if m3u8_match:
    print("Found M3U8:", m3u8_match.group(1))
else:
    print("No M3U8 found")
