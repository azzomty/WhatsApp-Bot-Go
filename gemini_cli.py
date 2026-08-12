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
        if response and hasattr(response, 'text'):
            print(response.text)
        elif response and hasattr(response, 'payload') and hasattr(response.payload, 'text'):
            print(response.payload.text)
        else:
            print("Error: Could not parse response.")
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    main()
