# Description
Application for compressing images of two types(.png and jpg/jpeg). It's consist of two microservices linked between each other by queue(RabbitMQ). Uploaded image saves in OS and after compressing will be rewrited.

# Usage
For running app, you must change constants of paths to suitable for you. And also running docker's image "rabbitmq" by this command:

		docker run -d --hostname my-rabbit --name some-rabbit -p 15672:15672 -p 5672:5672 rabbitmq:3-management
		
There is a index.html for comfortable uploading pictures. Now you can run two microservices for 
