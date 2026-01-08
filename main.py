
import http.client

def main():
    connection = http.client.HTTPConnection('www.python.org', 80, timeout=10)
    print(connection)

if __name__=="__main__":
    main()