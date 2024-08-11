import json
import pika
import pika.spec
from gridfs import GridFS
from typing import Dict, Tuple


def upload(f, fs: GridFS, channel, access: Dict) -> Tuple[str, int]:
	"""
	Uploads a given video file to the MongoDB utilizing GridFS. If the upload succeeds a message containing the
	video's information is published to RabbitMQ's video queue. If anything fails, an internal server error is returned.

	Parameters
	- f: Video file
	- fs: GridFS instance for videos
	- channel: RabbitMQ channe
	- JWT

	Returns
	- (str, int): Message, status code
	"""
	try:
		fid = fs.put(f)
	except Exception as e:
		# Upload failed
		print(f'Upload failed with the following error:\n{e}')
		return 'Internal server error', 500

	message = {
		'video_fid': str(fid),
		'mp3_fid': None,
		'username': access['username'] # The username is unique
	}

	try:
		# Attempt to add the message to the queue (RabbitMQ)
		channel.basic_publish(
			exchange='',
			routing_key='video',
			body=json.dumps(message),
			properties=pika.BasicProperties(
				delivery_mode=pika.spec.PERSISTENT_DELIVERY_MODE # Make the queue retain its messages even if the pod is restarted
			)
		)
	except Exception as e:
		# Queueing failed, delete the file from MongoDB
		print(f'Queuing failed with the following error:\n{e}')
		fs.delete(fid)
		return 'Internal server error', 500
