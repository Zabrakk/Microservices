import json
import pika
import pika.spec
from gridfs import GridFS
from typing import Dict


def upload(f, fs: GridFS, channel, access: Dict):
	"""
	Uploads a given file to the MongoDB using GridFS.
	"""
	try:
		fid = fs.put(f)
	except Exception as e:
		# Upload failed
		print(e)
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
		print(e)
		fs.delete(fid)
		return 'Internal server error', 500
