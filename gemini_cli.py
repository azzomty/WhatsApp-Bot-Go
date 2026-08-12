import sys
from gemini import Gemini

def main():
    sys.stdout.reconfigure(encoding='utf-8')
    if len(sys.argv) < 4:
        print("Usage: gemini_cli <cookie_1psid> <cookie_1psidts> <prompt>")
        sys.exit(1)
    
    cookie_1psid = sys.argv[1]
    cookie_1psidts = sys.argv[2]
    prompt = sys.argv[3]
    
    cookies = {
        "__Secure-1PSID": cookie_1psid,
        "__Secure-1PSIDTS": cookie_1psidts
    }
    
    try:
        client = Gemini(cookies=cookies)
        response = client.generate_content(prompt)
        text_output = ""
        
        if response and hasattr(response, 'text'):
            text_output = response.text
        elif response and hasattr(response, 'payload') and hasattr(response.payload, 'text'):
            text_output = response.payload.text
        else:
            text_output = "Error: Could not parse response."
            
        print(text_output)
        
        # Check for images
        images = []
        if hasattr(response, 'generated_images') and response.generated_images:
            images = response.generated_images
        elif hasattr(response, 'candidates') and len(response.candidates) > 0 and hasattr(response.candidates[0], 'generated_images'):
            images = response.candidates[0].generated_images
            
        if images:
            print("---MEDIA---")
            for img in images:
                print(str(img.url))
                
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    main()
