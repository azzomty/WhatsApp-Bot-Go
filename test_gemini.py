import sys
from gemini import Gemini
cookies = {
    "__Secure-1PSID": "g.a000BgkyVrxC8cwsVYW5dLfLHd-KlWxi4gxc32Uv8-Ee8FKXA51i5udy6dsl2ETs9tvosAI61QACgYKATASARUSFQHGX2MivIgMfl2Oqgtdx_Ae5yGcQBoVAUF8yKrTAw1qZVJgsEXATsPto0cv0076",
    "__Secure-1PSIDTS": "sidts-CjEBPWEu2Qdx4paX5rcNpwudpzTgTl1EKLjkScswFeHZj8FjKmm0dgbCVl54AbeKfwDnEAA",
    "__Secure-1PSIDCC": "AKEyXzXKC-oKhMpOZ8e3mYtnJ-ShXiMOZLMrrkDJJ49zPPaJcPJm5Zoy3sPfmZ6sPFG3KkEQfZRm"
}
try:
    client = Gemini(cookies=cookies)
    resp = client.generate_content("generate an image of a cat")
    print(resp.text)
    print(dir(resp))
    if hasattr(resp, 'images'): print("images:", resp.images)
    if hasattr(resp, 'generated_images'): print("generated_images:", resp.generated_images)
except Exception as e:
    print(e)
