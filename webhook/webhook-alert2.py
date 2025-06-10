from flask import Flask, request
import json

app = Flask(__name__)

@app.route('/', methods=['POST'])
def handle_webhook():
    data = request.get_json()
    print("接收到的告警数据:\n", json.dumps(data, indent=2))
    return "OK", 200

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8080)
