from flask import Flask, request, jsonify
import boto3
import os
from botocore.exceptions import ClientError

app = Flask(__name__)

# Initialize DynamoDB client
dynamodb = boto3.resource('dynamodb', region_name='us-west-2')  # Change region if needed
table = dynamodb.Table('HelloWorldTable')

@app.route('/')
def hello_world():
    return 'Hello, World!'

@app.route('/item', methods=['POST'])
def create_item():
    data = request.json
    try:
        response = table.put_item(Item=data)
        return jsonify({"message": "Item created successfully"}), 201
    except ClientError as e:
        return jsonify({"error": str(e)}), 400

@app.route('/item/<string:id>', methods=['GET'])
def get_item(id):
    try:
        response = table.get_item(Key={'ID': id})
        item = response.get('Item')
        if item:
            return jsonify(item)
        else:
            return jsonify({"message": "Item not found"}), 404
    except ClientError as e:
        return jsonify({"error": str(e)}), 400

if __name__ == '__main__':
    app.run(debug=True)