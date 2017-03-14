#!/usr/bin/python
# encoding=utf-8

# 先执行 python main.py 然后运行ggbot 会收到login msg contact 的消息推送
from BaseHTTPServer import BaseHTTPRequestHandler,HTTPServer

class WebhookHandler(BaseHTTPRequestHandler):
    def do_POST(self):

        content_length = int(self.headers['Content-Length']) # <--- Gets the size of data
        post_data = self.rfile.read(content_length) # <--- Gets the data itself

        print self.path
        print post_data

#Create a web server and define the handler to manage the
#incoming request
server = HTTPServer(('', 3288), WebhookHandler)

print 'Started httpserver on port 3288'

#Wait forever for incoming http requests
server.serve_forever()
