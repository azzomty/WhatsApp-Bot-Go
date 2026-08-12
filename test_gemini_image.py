import sys
from gemini import Gemini

cookies = {
    "__Secure-1PSID": "g.a000BgkyVrxC8cwsVYW5dLfLHd-KlWxi4gxc32Uv8-Ee8FKXA51i5udy6dsl2ETs9tvosAI61QACgYKATASARUSFQHGX2MivIgMfl2Oqgtdx_Ae5yGcQBoVAUF8yKrTAw1qZVJgsEXATsPto0cv0076",
    "__Secure-1PSIDTS": "sidts-CjEBPWEu2XHmxY-HfLlcBIfHKBw-4VRrbeyhKEIUv87IgE2p0KtL0uLMwSUi2xWV51NgEAA"
}

try:
    client = Gemini(cookies=cookies)
    response = client.generate_content("Generate an image of a cat riding a skateboard")
    print("Has payload?", hasattr(response, 'payload'))
    if hasattr(response, 'images'):
        print("Images directly:", response.images)
    elif hasattr(response, 'payload') and hasattr(response.payload, 'images'):
        print("Images in payload:", response.payload.images)
    elif hasattr(response, 'candidates'):
        print("Has candidates?", len(response.candidates))
        if hasattr(response.candidates[0], 'images'):
            print("Images in candidate:", response.candidates[0].images)
    print("Text:", response.text)
except Exception as e:
    print("Error:", e)
