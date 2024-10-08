import os
from typing import Tuple
import pika
import gridfs
from flask import Flask, request, send_file, Response
from flask_pymongo import PyMongo
from bson.objectid import ObjectId
from bson.errors import InvalidId

from auth_service import access, validation
from storage import util

server = Flask(__name__)

# MongoDB configurations
mongo_videos = PyMongo(server, uri=f'mongodb://{os.getenv("MONGODB_HOST")}:{os.getenv("MONGODB_PORT")}/videos')
mongo_mp3s = PyMongo(server, uri=f'mongodb://{os.getenv("MONGODB_HOST")}:{os.getenv("MONGODB_PORT")}/mp3s')

# GridFS for MongoDB file uploads / downloads
fs_videos = gridfs.GridFS(mongo_videos.db)
fs_mp3s = gridfs.GridFS(mongo_mp3s.db)


def create_connection() -> pika.BlockingConnection:
	"""
	Creates a connection to the RabbitMQ instance. Using this function ensures that there is always
	a working connection to RabbitMQ when files are being uploaded.

	Returns
	- BlockingConnection
	"""
	return pika.BlockingConnection(pika.ConnectionParameters(
		host="rabbitmq"
	))


@server.route('/login', methods=['POST'])
def login() -> str | Tuple[str, int]:
	"""
	Forwards the login request to the Authorization service and returns the service's response.
	Flask sets the status code to 200 with the token return.

	Returns
	- JWT or error (msg, status code)
	"""
	token, err = access.login(request)
	if not err:
		return token
	return err


@server.route('/register', methods=['POST'])
def register() -> str | Tuple[str, int]:
	"""
	Forwards the registration request to the Authorization service and returns the service's response.
	Flask sets the status code to 200 with the token return.

	Returns
	- JWT or error (msg, status code)
	"""
	token, err = access.register_user(request)
	if not err:
		return token
	return err


@server.route('/upload', methods=['POST'])
def upload() -> Tuple[str, int]:
	"""
	Uploads an authorized user's video to MongoDB and publishes a message about it on the RabbitMQ video queue.

	Returns
	- (str, int): Message, status code
	"""
	access, err = validation.validate_token(request)
	if not access:
		return err

	if access['admin']:
		if len(request.files) != 1:
			return 'Exactly one file required', 400

		connection = create_connection()
		for _, f in request.files.items():
			err = util.upload(f, fs_videos, connection.channel(), access)
			connection.close()
			if err:
				return err
		return 'Success!', 200
	else:
		return 'Not authorized', 401


@server.route('/download', methods=['GET'])
def download() -> Response | Tuple[str, int]:
	"""
	Downloads an audio file from MongoDB based on the FID, provided by an authorized user in their request's arguments.

	Returns
	- Response or (str, int): Audio file or error message and status code
	"""
	access, err = validation.validate_token(request)
	if not access:
		return err

	if access['admin']:
		fid_string = request.args.get('fid')
		if not fid_string:
			return 'URL parameter "fid" is required', 400
		try:
			out = fs_mp3s.get(ObjectId(fid_string))
			return send_file(out, download_name=f'{fid_string}.mp3')
		except gridfs.NoFile:
			return f'No file found with fid: {fid_string}', 404
		except InvalidId:
			return f'fid format is incorrect', 400
		except Exception as e:
			print(f'An error occured while trying to send file to user:\n{e}')
			return 'Internal server error', 500
	return 'Not authorized', 401


if __name__ == '__main__':
	server.run(host='0.0.0.0', port=8080)
