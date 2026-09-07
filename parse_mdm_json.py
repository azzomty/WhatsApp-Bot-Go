import re
import json

with open("mdm.html", "r") as f:
    content = f.read()

# SvelteKit embeds data typically in <script type="application/json" data-sveltekit-fetched> or similar
# Let's search for json blocks.
blocks = re.findall(r'<script.*?>(\{.*?\})</script>', content, re.DOTALL)
for b in blocks:
    if "Tier" in b:
        print("Found JSON block with Tier!")
        break
else:
    print("No JSON block found. It might be in the JS chunks.")
