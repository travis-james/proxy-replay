
import requests

def main():
    resp = requests.get('https://httpbin.org/get')
    print("Status code: {}, Resp JSON: {}".format(resp.status_code, resp.json()))

if __name__=="__main__":
    main()