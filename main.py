
import requests

def send_request(url: str):
    resp = requests.get(url)
    print("Status code: {}, Resp JSON: {}".format(resp.status_code, resp.json()))

def main():
    send_request('https://httpbin.org/get')

if __name__=="__main__":
    main()