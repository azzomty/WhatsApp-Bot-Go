from gemini import Gemini

cookies = {
    "__Secure-1PSID": "g.a000BgkyVrxC8cwsVYW5dLfLHd-KlWxi4gxc32Uv8-Ee8FKXA51i5udy6dsl2ETs9tvosAI61QACgYKATASARUSFQHGX2MivIgMfl2Oqgtdx_Ae5yGcQBoVAUF8yKrTAw1qZVJgsEXATsPto0cv0076",
    "__Secure-1PSIDTS": "sidts-CjEBPWEu2XHmxY-HfLlcBIfHKBw-4VRrbeyhKEIUv87IgE2p0KtL0uLMwSUi2xWV51NgEAA"
}

try:
    client = Gemini(cookies=cookies)
    response = client.generate_content("Hello")
    print(response.text)
except Exception as e:
    print(f"Error: {e}")
