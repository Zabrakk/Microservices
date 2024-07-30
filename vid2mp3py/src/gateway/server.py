import os
import json
import pika
import gridfs
from flask import Flask, request, send_file
from flask_pymongo import PyMongo
from bson.objectid import ObjectId

from auth import validate
from auth_svc import access
from storage import util

server = Flask(__name__)

mongo_videos = PyMongo(server, uri=f'mongodb://{os.getenv("MONGODB_HOST")}:{os.getenv("MONGODB_PORT")}/videos')
mongo_mp3s = PyMongo(server, uri=f'mongodb://{os.getenv("MONGODB_HOST")}:{os.getenv("MONGODB_PORT")}/mp3s')

fs_videos = gridfs.GridFS(mongo_videos.db)
fs_mp3s = gridfs.GridFS(mongo_mp3s.db)

connection = pika.BlockingConnection(pika.ConnectionParameters(
	host="rabbitmq"
))
channel = connection.channel()


@server.route("/login", methods=['POST'])
def login():
	token, err = access.login(request)
	if not err:
		return token
	else:
		return err


@server.route('/upload', methods=['POST'])
def upload():
	access, err = validate.token(request)
	if not access:
		return err

	if access['admin']:
		if len(request.files) != 1:
			return 'Exactly one file required', 400

		for _, f in request.files.items():
			err = util.upload(f, fs_videos, channel, access)
			if err:
				return err
		return 'Success!', 200
	else:
		return 'Not authorized', 401


@server.route('/download', methods=['GET'])
def download():
	access, err = validate.token(request)
	if not access:
		return err

	if access['admin']:
		fid_string = request.args.get('fid')
		if not fid_string:
			return 'URL parameter "fid" is required', 400
		try:
			out = fs_mp3s.get(ObjectId(fid_string))
			return send_file(out, download_name=f'{fid_string}.mp3')
		except Exception as e:
			print(f'An error occured while trying to send file to user:\n{e}')
			return 'Internal server error', 500
	return 'Not authorized', 401


if __name__ == '__main__':
	server.run(host='0.0.0.0', port=8080)
