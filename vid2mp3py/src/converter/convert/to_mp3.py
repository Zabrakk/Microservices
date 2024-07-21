import os
import json
import pika
import tempfile
import moviepy.editor
from bson.objectid import ObjectId
import pika.spec


def start(message, fs_videos, fs_mp3s, channel):
	message = json.loads(message)

	# Emptry temp file
	tf = tempfile.NamedTemporaryFile()

	# Get video contents
	out = fs_videos.get(ObjectId(message['video_fid']))

	# Add video contents to emptry file
	tf.write(out.read())

	# Get audio from the temp file
	audio = moviepy.editor.VideoFileClip(tf.name).audio
	tf.close()

	# Write audio to file
	tf_path = tempfile.gettempdir() + f'/{message["video_fid"]}.mp3'
	audio.write_audiofile(tf_path)

	# Save the audio file to MongoDB
	f = open(tf_path, 'rb')
	data = f.read()
	fid = fs_mp3s.put(data)
	f.close()
	os.remove(tf_path)

	# Update message mp3_fid
	message['mp3_fid'] = str(fid)

	# Put message on a new RabbitMQ queue
	try:
		channel.basic_publish(
			exchange='',
			routing_key=os.getenv('MP3_QUEUE'),
			body=json.dumps(message),
			properties=pika.BasicProperties(
				delivery_mode=pika.spec.PERSISTENT_DELIVERY_MODE
			)
		)
		return None
	except:
		# If we can't add the message to the queue, delete the file from MongoDB
		fs_mp3s.delete(fid)
		return 'Failed to publish message to MP3_QUEUE'
